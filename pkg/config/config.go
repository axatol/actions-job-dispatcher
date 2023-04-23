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
		LogLevel  *string        `yaml:"log_level"`
		LogFormat *string        `yaml:"log_format"`
		Interval  *string        `yaml:"interval"`
		Runners   []RunnerConfig `yaml:"runners"`
	}

	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		panic(fmt.Errorf("could not parse config file: %s", err))
	}

	if !LogLevel.set {
		if err := LogLevel.MaybeSet(cfg.LogLevel); err != nil {
			panic(fmt.Errorf("error loading log level: %s", err))
		}
	}

	if !LogFormat.set {
		if err := LogFormat.MaybeSet(cfg.LogFormat); err != nil {
			panic(fmt.Errorf("error loading log format: %s", err))
		}
	}

	if !Interval.set {
		if err := Interval.MaybeSet(cfg.Interval); err != nil {
			panic(fmt.Errorf("error loading interval: %s", err))
		}
	}

	Runners = cfg.Runners
	for i, runner := range Runners {
		if err := runner.Validate(); err != nil {
			panic(fmt.Errorf("failed to validate runner %d: %s", i, err))
		}
	}
}
