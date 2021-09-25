# Chronos

Chronos (previously known as WaybackUnifier) extracts pieces of data from a web page's history. It can be used to create custom wordlists, search for secrets, find old endpoints, etc.

---

## Installation
### From binary
Download a prebuilt binary from the [releases page](https://github.com/mhmdiaa/chronos/releases/latest) and unzip it.

### From source
Use `go get` to download and install the latest version
```
go get -u github.com/mhmdiaa/chronos
```

---

## Presets
Presets are predefined options (URL path, match regex, and extract regex) that can be used to organize and simplify different use cases. The preset definitions are stored in `~/.chronos` as JSON files

```
$ cat ~/.chronos/robots.json
{
        "path": "/robots.txt",
        "match": "Disallow",
        "extract": "(?:\\s)(/.*)"
}
$ chronos -pr robots -t example.com
$ # equivalent to...
$ chronos -p /robots.txt -m Disallow -e "(?:\\s)(/.*)"
```

---

## Example usage

### Extract paths from robots.txt files and build a wordlist
```
$ chronos -t example.com -p /robots.txt -m Disallow -e "(?:\\s)(/.*)" -o robots_wordlist.txt
```

### Save all versions of a web page locally and filter out a specifc status code
```
$ chronos -t http://example.com/this_is_403_now_but_has_it_always_been_like_this_question_mark -fs 403 -od output
```

### Save URLs of all subdomains of example.com that were last seen in 2015
```
$ chronos -t *.example.com -u -to 2015
```

### Run the S3 preset that extract AWS S3 URLs
```
$ chronos -pr s3 -t example.com
```


---

## Options
```
Usage: chronos <preset (optional)> <params>
  -c int
    	Number of concurrent threads (default 10)
  -e string
    	Extract regex
  -fm string
    	Filter Mime codes
  -from string
    	Match results after a specific date (Format: yyyyMMddhhmmss)
  -fs string
    	Filter status codes
  -m string
    	Match regex
  -mm string
    	Match Mime codes
  -ms string
    	Match status codes (default "200")
  -o string
    	Output file path (default "output.txt")
  -od string
    	Directory path to store matched results' entire pages
  -p string
    	Path to add to the URL
  -preset string
    	Preset name
  -t string
    	Target URL/domain (supports wildcards)
  -to string
    	Match results before a specific date (Format: yyyyMMddhhmmss)
  -u	URLs only
```

---

## Contributing
Find a bug? Got a feature request? Have an interesting preset in mind? Issues and pull requests are always welcome :)