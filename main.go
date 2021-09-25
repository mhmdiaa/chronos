package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type config struct {
	Target       string
	MatchRegex   string
	ExtractRegex string
	Path         string
	Concurrency  int
	URLsOnly     bool

	Filters filters

	OutputFile      string
	outputDirectory string
}

type filters struct {
	From             string
	To               string
	StatusMatchList  string
	StatusFilterList string
	MimeMatchList    string
	MimeFilterList   string
}

type preset struct {
	Path         string `json:"path"`
	MatchRegex   string `json:"match"`
	ExtractRegex string `json:"extract"`
}

func main() {
	flag.Usage = usage
	config := CreateConfig()

	if config.Target == "" {
		usage()
		os.Exit(1)
	}

	// If a path is provided, append it to the target
	if config.Path != "" {
		config.Target = strings.TrimRight(config.Target, "/") + config.Path
	}
	fmt.Printf("[*] Target: %s\n", config.Target)

	snapshots, err := getListOfSnapshots(config.Target, config.Filters, config.URLsOnly)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[*] Found %d snapshots\n", len(snapshots))

	if config.URLsOnly {
		err := writeURLsToFile(config.OutputFile, snapshots)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if config.outputDirectory != "" {
		err = os.MkdirAll(config.outputDirectory, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Global variable where the results are collected from all concurrent goroutines
	allResults := struct {
		sync.RWMutex
		results map[string]bool
	}{results: make(map[string]bool)}

	limit := make(chan bool, config.Concurrency)

	for _, snapshot := range snapshots {
		limit <- true
		go func(snapshot []string) {
			defer func() { <-limit }()
			fmt.Printf("[*] Requesting snapshot %s/%s\n", snapshot[0], snapshot[1])
			results, err := MatchAndExtractFromSnapshot(snapshot, config.MatchRegex, config.ExtractRegex, config.outputDirectory)

			if err != nil {
				fmt.Printf("[X] %v\n", err)
			} else {
				fmt.Printf("[*] Found %d matches in %s/%s\n", len(results), snapshot[0], snapshot[1])
				for _, result := range results {
					allResults.Lock()
					allResults.results[result] = true
					allResults.Unlock()
				}
			}
		}(snapshot)
	}
	err = writeSetToFile(config.OutputFile, allResults.results)
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Println("Usage: chronos <preset (optional)> <params>")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println()
	fmt.Println("EXAMPLE USAGE")
	fmt.Println("  Extract paths from robots.txt files and build a wordlist")
	fmt.Println("    chronos -t example.com -p /robots.txt -m Disallow -e \"(?:\\s)(/.*)\" -o robots_wordlist.txt")
	fmt.Println()
	fmt.Println("  Save all versions of a web page locally and filter out a specifc status code")
	fmt.Println("    chronos -t http://example.com/this_is_403_now_but_has_it_always_been_like_this_question_mark -fs 403 -od output")
	fmt.Println()
	fmt.Println("  Save URLs of all subdomains of example.com that were last seen in 2015")
	fmt.Println("    chronos -t *.example.com -u -to 2015")
}

func CreateConfig() config {
	var c config

	// General options
	flag.StringVar(&c.Target, "t", "", "Target URL/domain (supports wildcards)")
	flag.StringVar(&c.MatchRegex, "m", "", "Match regex")
	flag.StringVar(&c.ExtractRegex, "e", "", "Extract regex")
	flag.StringVar(&c.Path, "p", "", "Path to add to the URL")
	flag.IntVar(&c.Concurrency, "c", 10, "Number of concurrent threads")
	flag.BoolVar(&c.URLsOnly, "u", false, "URLs only")
	flag.StringVar(&c.OutputFile, "o", "output.txt", "Output file path")
	flag.StringVar(&c.outputDirectory, "od", "", "Directory path to store matched results' entire pages")
	var presetName string
	flag.StringVar(&presetName, "preset", "", "Preset name")

	// Filter options
	flag.StringVar(&c.Filters.From, "from", "", "Match results after a specific date (Format: yyyyMMddhhmmss)")
	flag.StringVar(&c.Filters.To, "to", "", "Match results before a specific date (Format: yyyyMMddhhmmss)")
	flag.StringVar(&c.Filters.StatusMatchList, "ms", "200", "Match status codes")
	flag.StringVar(&c.Filters.StatusFilterList, "fs", "", "Filter status codes")
	flag.StringVar(&c.Filters.MimeMatchList, "mm", "", "Match Mime codes")
	flag.StringVar(&c.Filters.MimeFilterList, "fm", "", "Filter Mime codes")

	flag.Parse()

	// Overwrite the parameters if a preset is passed
	if presetName != "" {
		preset, err := parsePreset(presetName)
		if err != nil {
			fmt.Printf("[X] Error loading the preset: %v\n", err)
			fmt.Println("[X] Preset will be ignored")
		} else {
			if preset.MatchRegex != "" {
				c.MatchRegex = preset.MatchRegex
			}
			if preset.ExtractRegex != "" {
				c.ExtractRegex = preset.ExtractRegex
			}
			if preset.ExtractRegex != "" {
				c.Path = preset.Path
			}
		}
	}
	return c
}

func getPresetDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	path := filepath.Join(usr.HomeDir, ".chronos")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path, nil
	} else {
		return "", err
	}
}

func parsePreset(presetName string) (preset, error) {
	presetDir, err := getPresetDir()
	if err != nil {
		return preset{}, err
	}

	presetFile := filepath.Join(presetDir, presetName+".json")

	var p preset
	file, err := ioutil.ReadFile(presetFile)
	if err != nil {
		return preset{}, err
	}
	err = json.Unmarshal(file, &p)
	if err != nil {
		return preset{}, err
	}
	return p, nil
}

func ConvertCommaSeparatedListToURLParams(list string, filter string, negative bool) string {
	params := ""
	var filterParam string
	if negative {
		filterParam = "&filter=!%s:%s"
	} else {
		filterParam = "&filter=%s:%s"
	}

	if list != "" {
		for _, item := range strings.Split(list, ",") {
			params += fmt.Sprintf(filterParam, filter, item)
		}
	}
	return params
}

// Search WaybackMachine for a given URL and returns a slice of [timestamp, url] slices
func getListOfSnapshots(target string, filters filters, removeDuplicateURLs bool) ([][]string, error) {
	searchURL := "https://web.archive.org/cdx/search/cdx?output=json"
	searchURL += "&url=" + target
	if filters.From != "" {
		searchURL += "&from=" + filters.From
	}
	if filters.To != "" {
		searchURL += "&to=" + filters.To
	}
	searchURL += ConvertCommaSeparatedListToURLParams(filters.StatusMatchList, "statuscode", false)
	searchURL += ConvertCommaSeparatedListToURLParams(filters.StatusFilterList, "statuscode", true)
	searchURL += ConvertCommaSeparatedListToURLParams(filters.MimeMatchList, "mimetype", false)
	searchURL += ConvertCommaSeparatedListToURLParams(filters.MimeFilterList, "mimetype", true)
	if removeDuplicateURLs {
		searchURL += "&collapse=urlkey&fl=original"
	} else {
		searchURL += "&collapse=digest&fl=timestamp,original"
	}

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("Error while loading search results: %v", err)
	}
	defer resp.Body.Close()

	var results [][]string
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error while reading search results: %v", err)
	}

	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, fmt.Errorf("Error while deserializing search results: %v", err)
	}
	if len(results) == 0 {
		return [][]string{}, fmt.Errorf("Didn't find any WaybackMachine entries for %s", target)
	}

	// The first item in the list is metadata
	return results[1:], nil
}

// Get the content of a WaybackMachine snapshot and return it as a string
func getContentOfSnapshot(snapshot []string) (string, error) {
	timestamp := snapshot[0]
	original := snapshot[1]

	url := "https://web.archive.org/web/" + timestamp + "if_" + "/" + original
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Error while loading a snapshot %s/%s: %v", timestamp, original, err)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error while reading a snapshot %s/%s: %v", timestamp, original, err)
	}
	return string(content), nil
}

func MatchAndExtractFromSnapshot(snapshot []string, mRegex, eRegex, outputDir string) ([]string, error) {
	content, err := getContentOfSnapshot(snapshot)
	if err != nil {
		fmt.Println(err)
	}

	if matchRegex(content, mRegex) {
		if outputDir != "" {
			filePath := filepath.Join(outputDir, snapshot[0]+"_"+strings.Replace(snapshot[1], "/", "_", -1))
			err = writeContentToFile(filePath, content)
			if err != nil {
				fmt.Println(err)
			}
		}
		results := extractRegex(content, eRegex)
		return results, nil
	} else {
		return []string{}, nil
	}
}

// Returns true if a page's content contains a string that matches the refex
func matchRegex(content, regex string) bool {
	r := regexp.MustCompile(regex)
	matches := r.Find([]byte(content))

	if matches != nil {
		return true
	}
	return false
}

// Extract all matches of a regex from a page's content
func extractRegex(content, regex string) []string {
	r := regexp.MustCompile(regex)
	var paths []string
	matches := r.FindAllStringSubmatch(string(content), -1)
	for _, i := range matches {
		paths = append(paths, i[1])
	}
	return paths
}

func writeURLsToFile(filePath string, values [][]string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, value := range values {
		fmt.Fprintln(f, value[0])
	}
	return nil
}

func writeSetToFile(filePath string, values map[string]bool) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	for value := range values {
		fmt.Fprintln(f, value)
	}
	return nil
}

func writeContentToFile(filePath string, content string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, content)
	if err != nil {
		return err
	}
	return nil
}
