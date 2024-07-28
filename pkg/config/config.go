package config

import (
	"flag"
	"fmt"

	"github.com/mhmdiaa/chronos/pkg/wayback"
)

type Config struct {
	Target        string
	Modules       string
	ModuleOptions moduleOptions
	ListModules   bool
	Filters       wayback.Filters
	Threads       int
	ConfigFile    string
	BaseURL       string
	OutputFile    string
}

type moduleOptions []string

func (moduleOptions *moduleOptions) String() string {
	return fmt.Sprintf("%s", *moduleOptions)
}

func (moduleOptions *moduleOptions) Set(value string) error {
	*moduleOptions = append(*moduleOptions, value)
	return nil
}

func NewConfig() Config {
	var c Config

	c.BaseURL = "https://web.archive.org"

	// General options
	flag.StringVar(&c.Target, "target", "", "Specify the target URL or domain (supports wildcards)")
	flag.StringVar(&c.Modules, "module", "", "Comma-separated list of modules to run")
	flag.Var(&c.ModuleOptions, "module-config", "Module configuration in the format: module.key=value")
	flag.BoolVar(&c.ListModules, "list-modules", false, "List available modules")
	flag.IntVar(&c.Threads, "threads", 10, "Number of concurrent threads to use")
	flag.StringVar(&c.ConfigFile, "module-config-file", "", "Path to the module configuration file")
	flag.StringVar(&c.OutputFile, "output", "", "Path to the output file")

	// Filter options
	flag.StringVar(&c.Filters.From, "from", "", "Filter snapshots from a specific date (Format: yyyyMMddhhmmss)")
	flag.StringVar(&c.Filters.To, "to", "", "Filter snapshots to a specific date (Format: yyyyMMddhhmmss)")
	flag.StringVar(&c.Filters.StatusMatchList, "match-status", "200", "Comma-separated list of status codes to match")
	flag.StringVar(&c.Filters.StatusFilterList, "filter-status", "", "Comma-separated list of status codes to filter out")
	flag.StringVar(&c.Filters.MimeMatchList, "match-mime", "", "Comma-separated list of MIME types to match")
	flag.StringVar(&c.Filters.MimeFilterList, "filter-mime", "", "Comma-separated list of MIME types to filter out")
	flag.StringVar(&c.Filters.Limit, "limit", "-50", "Limit the number of snapshots to process (use negative numbers for the newest N snapshots, positive numbers for the oldest N results)")
	flag.StringVar(&c.Filters.Interval, "snapshot-interval", "", "The interval for getting at most one snapshot (possible values: h, d, m, y)")
	flag.BoolVar(&c.Filters.OnePerURL, "one-per-url", false, "Fetch one snapshot only per URL")

	flag.Parse()

	return c
}
