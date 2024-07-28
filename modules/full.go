package modules

import (
	"sync"

	"github.com/mhmdiaa/chronos/pkg/wayback"
)

type Full struct {
	*BaseModule
}

func init() {
	module := &Full{
		BaseModule: NewBaseModule("full", "Get the full content of snapshots"),
	}
	RegisterModule(module)
}

func (module *Full) Process(snapshotChannel <-chan wayback.Snapshot, outputChannel chan<- ModuleOutput, wg *sync.WaitGroup, config ModuleConfig) {
	defer wg.Done()
	for snapshot := range snapshotChannel {
		output := ModuleOutput{
			Module:      module.Name(),
			URL:         snapshot.OriginalURL,
			SnapshotURL: snapshot.SnapshotURL,
			Results:     snapshot.Content,
		}
		outputChannel <- output
	}
}
