package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var (
	// global

	configFile string
	DryRun     bool
	logLevel   logLevelValue
	logFormat  logFormatValue

	// server

	ServerPort int64

	// github

	Github GithubConfig

	// kubernetes

	KubeConfig  string
	KubeContext string
	Namespace   string

	// reconciler

	SyncInterval time.Duration
	Runners      RunnerConfigList

	// metadata

	PrintVersion bool
)

func LoadConfig() {
	fs := flagSet{flag.CommandLine}
	fs.StringVar(&configFile, "config", "", "path to config")
	fs.BoolVar(&DryRun, "dry-run", false, "dry run")
	fs.Var(&logLevel, "log-level", "log level")
	fs.Var(&logFormat, "log-format", "log format")
	fs.Int64Var(&ServerPort, "server-port", 8000, "server port")
	fs.StringVar(&Github.Token, "github-token", "", "github token")
	fs.Int64Var(&Github.AppID, "github-app-id", 0, "github app id")
	fs.Int64Var(&Github.AppInstallationID, "github-app-installation-id", 0, "github app installation id")
	fs.StringVar(&Github.AppPrivateKey, "github-app-private-key", "", "github app private key")
	fs.StringVar(&KubeConfig, "kube-config", KubeConfig, "path to the kubeconfig file")
	fs.StringVar(&KubeContext, "kube-context", KubeContext, "specific a kubernetes context")
	fs.StringVar(&Namespace, "namespace", "actions-runners", "specify a kubernetes namespace")
	fs.DurationVar(&SyncInterval, "sync-interval", time.Minute*5, "sync interval")
	fs.BoolVar(&PrintVersion, "version", false, "prints current version")

	// flags first priority
	flag.Parse()

	godotenv.Load()
	fs.LoadUnsetFromEnv()

	// config file lowest priority
	loadConfigFromFile()

	zerolog.SetGlobalLevel(zerolog.Level(logLevel))
	if logFormat == logFormatValue(textLogFormat) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
}

func loadConfigFromFile() {
	filenames := []string{configFile, "./config.yaml"}

	for _, filename := range filenames {
		if filename == "" {
			continue
		}

		if _, err := os.Stat(filename); err != nil {
			continue
		}

		raw, err := os.ReadFile(filename)
		if err != nil {
			panic(fmt.Errorf("could not read config file at %s: %s", filename, err))
		}

		// config you want to load from a file
		var cfg struct {
			Github  GithubConfig   `yaml:"github"`
			Runners []RunnerConfig `yaml:"runners"`
		}

		if err := yaml.Unmarshal(raw, &cfg); err != nil {
			panic(fmt.Errorf("could not parse config file at %s: %s", filename, err))
		}

		// check if set from cli
		if err := Github.Validate(); err != nil {
			// if not, set from config
			Github = cfg.Github
			if err := Github.Validate(); err != nil {
				panic(fmt.Errorf("failed to validate github: %s", err))
			}
		}

		Runners = cfg.Runners
		if err := Runners.Validate(); err != nil {
			panic(fmt.Errorf("failed to validate runners: %s", err))
		}
	}
}
