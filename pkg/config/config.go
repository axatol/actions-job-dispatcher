package config

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig() {
	registerGlobalFlags()
	registerKubernetesFlags()
	registerReconcilerFlags()
	registerGithubFlags()

	flag.Parse()
	loadConfigFromFile()
	configureLogger()
}

func findConfigFile() string {
	filenames := []string{ConfigFile, "./config.yaml"}

	for _, filename := range filenames {
		if filename == "" {
			continue
		}

		if _, err := os.Stat(filename); err != nil {
			continue
		}

		return filename
	}

	return ""
}

func loadConfigFromFile() {
	filename := findConfigFile()
	if filename == "" {
		return
	}

	raw, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Errorf("could not read config file: %s", err))
	}

	var cfg struct {
		Runners []RunnerConfig `yaml:"runners"`
	}

	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		panic(fmt.Errorf("could not parse config file: %s", err))
	}

	Runners = cfg.Runners
	for i, runner := range Runners {
		if err := runner.Validate(); err != nil {
			panic(fmt.Errorf("failed to validate runner %d: %s", i, err))
		}
	}
}
