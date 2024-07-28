package modules

import (
	"bytes"
	"encoding/base64"
	"sync"

	"github.com/mhmdiaa/chronos/v2/pkg/logger"
	"github.com/mhmdiaa/chronos/v2/pkg/wayback"
	"github.com/spaolacci/murmur3"
)

type Favicon struct {
	*BaseModule
}

func init() {
	module := &Favicon{
		BaseModule: NewBaseModule("favicon", "Calculate favicon hashes"),
	}
	RegisterModule(module)
}

func (module *Favicon) Process(snapshotChannel <-chan wayback.Snapshot, outputChannel chan<- ModuleOutput, wg *sync.WaitGroup, config ModuleConfig) {
	defer wg.Done()
	for snapshot := range snapshotChannel {
		result := murmurhash([]byte(snapshot.Content))

		if result == 142044466 {
			logger.Info.Printf("[favicon] Skipped snapshot %s: points to Wayback Machine's own favicon", snapshot.SnapshotURL)
			continue
		}

		output := ModuleOutput{
			Module:      module.Name(),
			URL:         snapshot.OriginalURL,
			SnapshotURL: snapshot.SnapshotURL,
			Results:     result,
		}
		outputChannel <- output
	}
}

func murmurhash(data []byte) int32 {
	stdBase64 := base64.StdEncoding.EncodeToString(data)
	stdBase64 = InsertInto(stdBase64, 76, '\n')
	hasher := murmur3.New32WithSeed(0)
	hasher.Write([]byte(stdBase64))
	return int32(hasher.Sum32())
}

func InsertInto(s string, interval int, sep rune) string {
	var buffer bytes.Buffer
	before := interval - 1
	last := len(s) - 1
	for i, char := range s {
		buffer.WriteRune(char)
		if i%interval == before && i != last {
			buffer.WriteRune(sep)
		}
	}
	buffer.WriteRune(sep)
	return buffer.String()
}
