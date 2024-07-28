package modules

import (
	"os"
	"strings"
	"sync"

	"github.com/mhmdiaa/chronos/pkg/logger"
	"github.com/mhmdiaa/chronos/pkg/wayback"
	"gopkg.in/yaml.v2"
)

var ModuleRegistry = make(map[string]Module)

func RegisterModule(module Module) {
	ModuleRegistry[module.Name()] = module
}

type Module interface {
	Name() string
	Description() string
	Process(snapshotChannel <-chan wayback.Snapshot, outputChannel chan<- ModuleOutput, wg *sync.WaitGroup, config ModuleConfig)
	Channel() chan wayback.Snapshot
}

type ModuleOutput struct {
	Module      string      `json:"module,omitempty"`
	URL         string      `json:"url,omitempty"`
	SnapshotURL string      `json:"snapshot,omitempty"`
	Results     interface{} `json:"results,omitempty"`
}

type BaseModule struct {
	name        string
	description string
	channel     chan wayback.Snapshot
}

func (module *BaseModule) Name() string {
	return module.name
}

func (module *BaseModule) Description() string {
	return module.description
}

func (module *BaseModule) Channel() chan wayback.Snapshot {
	return module.channel
}

func NewBaseModule(name, description string) *BaseModule {
	return &BaseModule{
		name:        name,
		description: description,
		channel:     make(chan wayback.Snapshot),
	}
}

type ModuleConfig map[string]interface{}

func ParseModuleOptions(options []string) map[string]ModuleConfig {
	moduleOptions := make(map[string]ModuleConfig)
	for _, option := range options {
		parts := strings.SplitN(option, ".", 2)
		if len(parts) != 2 {
			logger.Error.Fatalf("invalid config format: %s", option)
		}
		moduleName := parts[0]
		keyValue := strings.SplitN(parts[1], "=", 2)
		if len(keyValue) != 2 {
			logger.Error.Fatalf("invalid config format: %s", option)
		}
		key := keyValue[0]
		value := keyValue[1]

		if _, exists := moduleOptions[moduleName]; !exists {
			moduleOptions[moduleName] = ModuleConfig{}
		}
		moduleOptions[moduleName][key] = value
	}
	return moduleOptions
}

func ParseModuleConfigFile(file string) map[string]ModuleConfig {
	moduleOptions := make(map[string]ModuleConfig)

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		logger.Error.Fatalf("failed to read the config file %s: %v", file, err)
	}

	config := make(map[string]string)
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		logger.Error.Fatalf("failed to parse the config file %s: %v", file, err)
	}

	for k, v := range config {
		parts := strings.SplitN(k, ".", 2)
		if len(parts) != 2 {
			logger.Error.Fatalf("invalid config format: %s", k)
		}
		moduleName := parts[0]
		key := parts[1]

		if _, exists := moduleOptions[moduleName]; !exists {
			moduleOptions[moduleName] = ModuleConfig{}
		}
		moduleOptions[moduleName][key] = v
	}
	return moduleOptions
}
