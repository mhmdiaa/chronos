package modules

import (
	"strings"
	"sync"

	"github.com/antchfx/xmlquery"
	"github.com/mhmdiaa/chronos/v2/pkg/logger"
	"github.com/mhmdiaa/chronos/v2/pkg/wayback"
)

type XML struct {
	*BaseModule
}

func init() {
	module := &XML{
		BaseModule: NewBaseModule("xml", "Query XML documents using XPath expressions"),
	}
	RegisterModule(module)
}

func (module *XML) Process(snapshotChannel <-chan wayback.Snapshot, outputChannel chan<- ModuleOutput, wg *sync.WaitGroup, config ModuleConfig) {
	defer wg.Done()

	for snapshot := range snapshotChannel {
		doc, err := xmlquery.Parse(strings.NewReader(snapshot.Content))
		if err != nil {
			logger.Warn.Printf("failed to parse %s as an HTML document", snapshot.SnapshotURL)
			continue
		}

		matches := make(map[string][]string)
		for label, expression := range config {
			var expressionMatches []string
			nodes, err := xmlquery.QueryAll(doc, expression.(string))
			if err != nil {
				logger.Warn.Printf("failed to run expression %s on %s", expression, snapshot.SnapshotURL)
				continue
			}
			for _, node := range nodes {
				value := node.InnerText()
				expressionMatches = append(expressionMatches, value)
			}
			if len(expressionMatches) > 0 {
				matches[label] = expressionMatches
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
