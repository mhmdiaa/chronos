package modules

import (
	"regexp"
	"sync"

	"github.com/mhmdiaa/chronos/v2/pkg/wayback"
)

type Regex struct {
	*BaseModule
}

func init() {
	module := &Regex{
		BaseModule: NewBaseModule("regex", "Extract regex matches"),
	}
	RegisterModule(module)
}

func (module *Regex) Process(snapshotChannel <-chan wayback.Snapshot, outputChannel chan<- ModuleOutput, wg *sync.WaitGroup, config ModuleConfig) {
	defer wg.Done()

	expressions := make(map[string]*regexp.Regexp)
	for label, expression := range config {
		expressions[label] = regexp.MustCompile(expression.(string))
	}

	for snapshot := range snapshotChannel {
		matches := make(map[string][]string)
		for label, re := range expressions {
			reMatches := re.FindAllString(snapshot.Content, -1)
			if len(reMatches) > 0 {
				matches[label] = reMatches
			}
		}

		if len(matches) > 0 {
			output := ModuleOutput{
				Module:      module.Name(),
				URL:         snapshot.OriginalURL,
				SnapshotURL: snapshot.SnapshotURL,
				Results:     matches,
			}
			outputChannel <- output
		}
	}
}
