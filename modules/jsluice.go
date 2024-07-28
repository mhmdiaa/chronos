package modules

import (
	"sync"

	"github.com/BishopFox/jsluice"
	"github.com/mhmdiaa/chronos/v2/pkg/wayback"
)

type JSLuice struct {
	*BaseModule
}

func init() {
	module := &JSLuice{
		BaseModule: NewBaseModule("jsluice", "Extract URLs and endpoints from JavaScript code using jsluice"),
	}
	RegisterModule(module)
}

func (module *JSLuice) Process(snapshotChannel <-chan wayback.Snapshot, outputChannel chan<- ModuleOutput, wg *sync.WaitGroup, config ModuleConfig) {
	defer wg.Done()

	for snapshot := range snapshotChannel {
		analyzer := jsluice.NewAnalyzer([]byte(snapshot.Content))
		urls := analyzer.GetURLs()

		if len(urls) > 0 {
			output := ModuleOutput{
				Module:      module.Name(),
				URL:         snapshot.OriginalURL,
				SnapshotURL: snapshot.SnapshotURL,
				Results:     urls,
			}
			outputChannel <- output
		}
	}
}
