package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var results = struct {
	sync.RWMutex
	res map[string]struct{}
}{res: make(map[string]struct{})}

func main() {
	url := flag.String("url", "site.com/robots.txt", "URL to unify versions of (without protocol prefix)")
	output := flag.String("output", "output.txt", "File to save results in")
	concurrency := flag.Int("concurrency", 1, "Number of requests to make in prallel")
	substrings := flag.String("sub", "Disallow,disallow", "list of comma-separated substrings to look for in snapshots (snapshots will only be considered if they contnain one of them)")

	flag.Parse()
	var subs []string

	for _, sub := range strings.Split(*substrings, ",") {
		subs = append(subs, sub)
	}

	snapshots, err := getSnapshots(*url)
	if err != nil {
		log.Fatalf("couldn't get snapshots: %v", err)
	}
	fmt.Printf("[*] Found %d snapshots", len(snapshots))

	lim := make(chan bool, *concurrency)
	for _, snapshot := range snapshots {
		lim <- true
		go func(snapshot []string) {
			defer func() { <-lim }()
			unifySnapshots(snapshot, subs)
			if err != nil {
				log.Printf("couldn't unify snapshots: %v", err)
			}
		}(snapshot)
	}

	for i := 0; i < cap(lim); i++ {
		lim <- true
	}

	r := ""
	for i := range results.res {
		r += i + "\n"
	}
	f, err := os.Create(*output)
	if err != nil {
		log.Fatalf("couldn't create output file: %v", err)
	}
	defer f.Close()

	f.Write([]byte(r))
}

func unifySnapshots(snapshot []string, subs []string) {
	content, err := getContent(snapshot)
	if err != nil {
		log.Printf("couldn't fetch snapshot: %v", err)
	}
	if len(subs) > 0 {
		foundSub := false
		for _, sub := range subs {
			if strings.Contains(content, sub) {
				foundSub = true
			}
		}
		if !foundSub {
			log.Printf("snapshot %s/%s doesn't contain any substring", snapshot[0], snapshot[1])
		}
	}
	c := strings.Split(content, "\n")
	for _, line := range c {
		results.Lock()
		if line != "" {
			results.res[line] = struct{}{}
		}
		results.Unlock()
	}
}

func getSnapshots(url string) ([][]string, error) {
	resp, err := http.Get("https://web.archive.org/cdx/search/cdx?url=" + url + "&output=json&fl=timestamp,original&filter=statuscode:200&collapse=digest")
	if err != nil {
		return nil, fmt.Errorf("coudln't load waybackmachine search results for %s: %v", url, err)
	}
	defer resp.Body.Close()

	var results [][]string
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("couldn't read waybackmachine search results for %s: %v", url, err)
	}

	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, fmt.Errorf("coudln't deserialize JSON response from waybackmachine for %s: %v", url, err)
	}
	if len(results) == 0 {
		return make([][]string, 0), fmt.Errorf("")
	}
	return results[1:], nil
}

func getContent(snapshot []string) (string, error) {
	timestamp := snapshot[0]
	original := snapshot[1]
	url := "https://web.archive.org/web/" + timestamp + "if_" + "/" + original
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("couldn't load snapshot for %s/%s: %v", timestamp, original, err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't read snapshot content for %s/%s: %v", timestamp, original, err)
	}
	return string(content), nil
}
