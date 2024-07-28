package wayback

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/mhmdiaa/chronos/v2/pkg/logger"
)

type Filters struct {
	From             string
	To               string
	StatusMatchList  string
	StatusFilterList string
	MimeMatchList    string
	MimeFilterList   string
	Limit            string
	Interval         string
	OnePerURL        bool
}

type Snapshot struct {
	OriginalURL string
	SnapshotURL string
	Content     string
}

func SearchForSnapshots(baseURL, target string, filters Filters) ([]Snapshot, error) {
	searchURL := buildSearchURL(baseURL, target, filters)

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get search results for %s: %v", target, err)
	}
	defer resp.Body.Close()

	return parseSearchResults(baseURL, resp.Body, target)
}

func buildSearchURL(baseURL, target string, filters Filters) string {
	searchURL := fmt.Sprintf("%s/cdx/search/cdx?output=json&fl=timestamp,original", baseURL)
	searchURL += "&url=" + target
	if filters.From != "" {
		searchURL += "&from=" + filters.From
	}
	if filters.To != "" {
		searchURL += "&to=" + filters.To
	}
	searchURL += formatFilterParams(filters.StatusMatchList, "statuscode", false)
	searchURL += formatFilterParams(filters.StatusFilterList, "statuscode", true)
	searchURL += formatFilterParams(filters.MimeMatchList, "mimetype", false)
	searchURL += formatFilterParams(filters.MimeFilterList, "mimetype", true)
	searchURL += formatFilterParams("warc/revisit", "mimetype", true)
	searchURL += "&limit=" + filters.Limit

	comparedDigits := getComparedDigits(filters.Interval)
	if comparedDigits != "" {
		searchURL += "&collapse=timestamp:" + comparedDigits
	} else if filters.OnePerURL {
		searchURL += "&collapse=urlkey"
	} else {
		searchURL += "&collapse=digest"
	}

	return searchURL
}

func getComparedDigits(interval string) string {
	switch interval {
	case "h":
		return "10"
	case "d":
		return "8"
	case "m":
		return "6"
	case "y":
		return "4"
	default:
		return ""
	}
}

func formatFilterParams(list string, filter string, negative bool) string {
	if list == "" {
		return ""
	}

	params := ""
	filterParam := "&filter=%s:%s"
	if negative {
		filterParam = "&filter=!%s:%s"
	}

	for _, item := range strings.Split(list, ",") {
		item := strings.ReplaceAll(item, "+", ".")
		params += fmt.Sprintf(filterParam, filter, item)
	}

	return params
}

func parseSearchResults(baseURL string, body io.Reader, target string) ([]Snapshot, error) {
	var results [][]string
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read search results for %s: %v", target, err)
	}

	err = json.Unmarshal(data, &results)
	if err != nil {
		fmt.Println(string(data))
		return nil, fmt.Errorf("failed to deserialize search results for %s: %v", target, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("found no snapshots of %s", target)
	}

	return convertResultsToSnapshots(baseURL, results), nil
}

func convertResultsToSnapshots(baseURL string, results [][]string) []Snapshot {
	var snapshots []Snapshot
	// The first item in the list is column names
	for _, s := range results[1:] {
		snapshot := Snapshot{
			OriginalURL: s[1],
			SnapshotURL: formatSnapshotResponseIntoURL(baseURL, s),
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots
}

func formatSnapshotResponseIntoURL(baseURL string, snapshot []string) string {
	timestamp := snapshot[0]
	original := snapshot[1]

	return fmt.Sprintf("%s/web/%sif_/%s", baseURL, timestamp, original)
}

func FetchSnapshots(snapshotLocations chan Snapshot, snapshots chan Snapshot, wg *sync.WaitGroup) {
	defer wg.Done()
	for location := range snapshotLocations {
		content, err := GetSnapshotContent(location.SnapshotURL)
		if err != nil {
			logger.Error.Print(err)
			continue
		}

		snapshot := Snapshot{
			OriginalURL: location.OriginalURL,
			SnapshotURL: location.SnapshotURL,
			Content:     content,
		}
		snapshots <- snapshot
	}
}

func GetSnapshotContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get snapshot %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read snapshot %s: %v", url, err)
	}
	content := string(body)

	return removeWaybackModifications(content), nil
}

func removeWaybackModifications(content string) string {
	waybackModifications := []string{
		`\<script type=\"text\/javascript" src=\"\/_static\/js\/bundle-playback\.js\?v=[A-Za-z0-9]*" charset="utf-8"><\/script>\n<script type="text\/javascript" src="\/_static\/js\/wombat\.js.*\<\!-- End Wayback Rewrite JS Include --\>`,
		`\<script src=\"\/\/archive\.org.*\<\!-- End Wayback Rewrite JS Include --\>`,
		`\<script\>window\.RufflePlayer[^\<]*\<\/script\>`,
		`\<\!-- BEGIN WAYBACK TOOLBAR INSERT --\>.*\<\!-- END WAYBACK TOOLBAR INSERT --\>`,
		`(}\n)?(\/\*|<!--\n)\s*FILE ARCHIVED ON.*108\(a\)\(3\)\)\.\n(\*\/|-->)`,
		`var\s_____WB\$wombat\$assign\$function.*WB\$wombat\$assign\$function_____\(\"opener\"\);`,
		`(\<\!--|\/\*)\nplayback timings.*(--\>|\*\/)`,
		`((https:)?\/\/web\.archive\.org)?\/web\/[0-9]{14}([A-Za-z]{2}\_)?\/`,
		`((https:)?\\\/\\\/web\.archive\.org)?\\\/web\\\/[0-9]{14}([A-Za-z]{2}\_)?\\\/`,
		`((https:)?%2F%2Fweb\.archive\.org)?%2Fweb%2F[0-9]{14}([A-Za-z]{2}\_)?%2F`,
		`((https:)?\\u002F\\u002Fweb\.archive\.org)?\\u002Fweb\\u002F[0-9]{14}([A-Za-z]{2}\_)?\\u002F`,
		`\<script type=\"text\/javascript\">\s*__wm\.init\(\"https:\/\/web\.archive\.org\/web\"\);[^\<]*\<\/script\>`,
		`\<script type=\"text\/javascript\" src="https:\/\/web-static\.archive\.org[^\<]*\<\/script\>`,
		`\<link rel=\"stylesheet\" type=\"text\/css\" href=\"https:\/\/web-static\.archive\.org[^\<]*\/\>`,
		`\<\!-- End Wayback Rewrite JS Include --\>`,
	}

	for _, mod := range waybackModifications {
		re := regexp.MustCompile(`(?i)` + mod)
		content = re.ReplaceAllString(content, "")
	}

	return content
}
