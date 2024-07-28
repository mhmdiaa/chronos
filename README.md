![chronos](./images/chronos.png)
# Chronos
Wayback Machine OSINT Framework

- [Installation](#installation)
- [Example Usage](#example-usage)
  - [Extract endpoints and URLs from archived JavaScript code](#extract-endpoints-and-urls-from-archived-javascript-code)
  - [Calculate archived favicon hashes](#calculate-archived-favicon-hashes)
  - [Extract archived page titles](#extract-archived-page-titles)
  - [Extract paths from archived robots.txt files](#extract-paths-from-archived-robotstxt-files)
  - [Extract URLs from archived sitemap.xml files](#extract-urls-from-archived-sitemapxml-files)
  - [Extract endpoints from archived API documentation](#enumerate-endpoints-from-api-documentation)
  - [Find S3 buckets in archived pages](#find-s3-buckets-in-archived-pages)
- [Modules](#modules)
- [Command-line Options](#command-line-options)

## Installation
### From binary
Download a prebuilt binary from the [releases page](https://github.com/mhmdiaa/chronos/releases/latest).

### From source
```
go install github.com/mhmdiaa/chronos@latest
```

## Example Usage
### Extract endpoints and URLs from archived JavaScript code
```
chronos -target "denali-static.grammarly.com/*" -module jsluice -output js_endpoints.json
```
[![asciicast](https://asciinema.org/a/lm8hSxIMWYk8f3wolSeNYJewO.svg)](https://asciinema.org/a/lm8hSxIMWYk8f3wolSeNYJewO)

### Calculate archived favicon hashes
```
chronos -target "netflix.com/favicon.ico" -module favicon -output favicon_hashes.json
```
[![asciicast](https://asciinema.org/a/sNKQA7XXnAFmOSKYUQyoph1vJ.svg)](https://asciinema.org/a/sNKQA7XXnAFmOSKYUQyoph1vJ)

### Extract archived page titles
```
chronos -target "github.com" -module html -module-config "html.title=//title" -snapshot-interval y -output titles.json
```
[![asciicast](https://asciinema.org/a/avNyaQoPN8WQ2vZkyrql7aYf0.svg)](https://asciinema.org/a/avNyaQoPN8WQ2vZkyrql7aYf0)

### Extract paths from archived robots.txt files
```
chronos -target "tripadvisor.com/robots.txt" -module regex -module-config 'regex.paths=/[^\s]+' -output robots_paths.json
```
[![asciicast](https://asciinema.org/a/zXw1XvhyNOIKPd41HWjWQvoIx.svg)](https://asciinema.org/a/zXw1XvhyNOIKPd41HWjWQvoIx)

### Extract URLs from archived sitemap.xml files
```
chronos -target "apple.com/sitemap.xml" -module xml -module-config "xml.urls=//urlset/url/loc" -limit -5 -output sitemap_urls.json
```
[![asciicast](https://asciinema.org/a/tJAWMuDx6z8G0pQqRCZR3WJiU.svg)](https://asciinema.org/a/tJAWMuDx6z8G0pQqRCZR3WJiU)

### Extract endpoints from archived API documentation
```
chronos -target "https://docs.gitlab.com/ee/api/api_resources.html" -module html -module-config 'html.endpoint=//code' -output api_docs_endpoints.json
```
[![asciicast](https://asciinema.org/a/5yrrAnt46CHJqlhja4T48ym8u.svg)](https://asciinema.org/a/5yrrAnt46CHJqlhja4T48ym8u)

### Find S3 buckets in archived pages
```
chronos -target "github.com" -module regex -module-config 'regex.s3=[a-zA-Z0-9-\.\_]+\.s3(?:-[-a-zA-Z0-9]+)?\.amazonaws\.com' -limit -snapshot-interval y -output s3_buckets.json
```
[![asciicast](https://asciinema.org/a/HKma8ycDMgHO6RPThjBXrOlrp.svg)](https://asciinema.org/a/HKma8ycDMgHO6RPThjBXrOlrp)

## Modules
| Module | Description                                                   |
|-------------|---------------------------------------------------------------|
| regex       | Extract regex matches                                         |
| jsluice     | Extract URLs and endpoints from JavaScript code using jsluice |
| html        | Query HTML documents using XPath expressions                  |
| xml         | Query XML documents using XPath expressions                   |
| favicon     | Calculate favicon hashes                                      |
| full        | Get the full content of snapshots                             |

## Command-line Options
```
Usage of chronos:
  -target string
    	Specify the target URL or domain (supports wildcards)
  -list-modules
    	List available modules
  -module string
    	Comma-separated list of modules to run
  -module-config value
    	Module configuration in the format: module.key=value
  -module-config-file string
    	Path to the module configuration file
  -match-mime string
    	Comma-separated list of MIME types to match
  -filter-mime string
    	Comma-separated list of MIME types to filter out
  -match-status string
    	Comma-separated list of status codes to match (default "200")
  -filter-status string
    	Comma-separated list of status codes to filter out
  -from string
    	Filter snapshots from a specific date (Format: yyyyMMddhhmmss)
  -to string
    	Filter snapshots to a specific date (Format: yyyyMMddhhmmss)
  -limit string
    	Limit the number of snapshots to process (use negative numbers for the newest N snapshots, positive numbers for the oldest N results) (default "-50")
  -snapshot-interval string
    	The interval for getting at most one snapshot (possible values: h, d, m, y)
  -one-per-url
    	Fetch one snapshot only per URL
  -threads int
    	Number of concurrent threads to use (default 10)
  -output string
    	Path to the output file
```
