package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/mhmdiaa/chronos/v2/modules"
	"github.com/mhmdiaa/chronos/v2/pkg/config"
	"github.com/mhmdiaa/chronos/v2/pkg/logger"
	"github.com/mhmdiaa/chronos/v2/pkg/wayback"
)

func main() {
	conf := config.NewConfig()
	err := logger.Init(conf.OutputFile)
	if err != nil {
		log.Fatalf("failed to create the output logger: %v", err)
	}
	defer logger.Close()

	if conf.ListModules {
		for _, module := range modules.ModuleRegistry {
			fmt.Println(module.Name())
			fmt.Printf("    %s\n", module.Description())
		}
		return
	}

	if conf.Target == "" {
		logger.Error.Fatal("target not specified")
	}

	logger.Info.Printf("Searching for snapshots...")
	snapshotLocationsList, err := wayback.SearchForSnapshots(conf.BaseURL, conf.Target, conf.Filters)
	if err != nil {
		logger.Error.Fatal(err)
	}
	logger.Info.Printf("Found %d snapshots\n", len(snapshotLocationsList))

	// If no modules are enabled, write snapshot locations and exit
	if conf.Modules == "" {
		for _, snapshot := range snapshotLocationsList {
			j, err := json.Marshal(snapshot)
			if err != nil {
				logger.Error.Println(err)
				continue
			}
			logger.Output.Println(string(j))
		}
		return
	}

	// Set up modules
	moduleNames := strings.Split(conf.Modules, ",")
	enabledModules := []modules.Module{}
	for _, moduleName := range moduleNames {
		if module, exists := modules.ModuleRegistry[moduleName]; exists {
			enabledModules = append(enabledModules, module)
		} else {
			logger.Error.Fatalf("Module %s not found", moduleName)
		}
	}

	var moduleWg sync.WaitGroup
	moduleWg.Add(len(enabledModules))
	outputChan := make(chan modules.ModuleOutput)

	var moduleConfig map[string]modules.ModuleConfig
	if conf.ConfigFile != "" {
		moduleConfig = modules.ParseModuleConfigFile(conf.ConfigFile)
	} else {
		moduleConfig = modules.ParseModuleOptions(conf.ModuleOptions)
	}

	for _, module := range enabledModules {
		config := moduleConfig[module.Name()]
		go module.Process(module.Channel(), outputChan, &moduleWg, config)
	}

	// Set up snapshot workers
	numOfWorkers := conf.Threads
	snapshotLocationsChan := make(chan wayback.Snapshot)
	snapshotsChan := make(chan wayback.Snapshot)

	var snapshotWg sync.WaitGroup
	snapshotWg.Add(numOfWorkers)

	for i := 0; i < numOfWorkers; i++ {
		go wayback.FetchSnapshots(snapshotLocationsChan, snapshotsChan, &snapshotWg)
	}

	go func() {
		for _, location := range snapshotLocationsList {
			snapshotLocationsChan <- location
		}
		close(snapshotLocationsChan)
	}()

	go func() {
		snapshotWg.Wait()
		close(snapshotsChan)
	}()

	// Connect the snapshot channel to the module channels
	// Then, close the module channels once the snapshots are completely processed
	go func() {
		for snapshot := range snapshotsChan {
			for _, module := range enabledModules {
				module.Channel() <- snapshot
			}
		}
		for _, module := range enabledModules {
			close(module.Channel())
		}
	}()

	// Close the output channel once all modules are done writing to it
	go func() {
		moduleWg.Wait()
		close(outputChan)
	}()

	for output := range outputChan {
		j, err := json.Marshal(output)
		if err != nil {
			logger.Error.Println(err)
			continue
		}
		logger.Output.Println(string(j))
	}
}
