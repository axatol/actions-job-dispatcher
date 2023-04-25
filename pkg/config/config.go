package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

var (
	// global

	ConfigFile string
	logLevel   = StringFlagValue{defaultValue: zerolog.InfoLevel.String(), validate: validateLogLevel}
	logFormat  = StringFlagValue{defaultValue: LogFormatJSON, validate: validateLogFormat}

	// github

	GithubToken             StringFlagValue
	GithubAppID             Int64FlagValue
	GithubAppInstallationID Int64FlagValue
	GithubAppPrivateKey     StringFlagValue

	// kubernetes

	kubeConfig  StringFlagValue
	kubeContext StringFlagValue
	Namespace   StringFlagValue

	// reconciler

	SyncInterval = IntFlagValue{defaultValue: 30, validate: validateSyncInterval}
)

func LoadConfig() {
	// global

	flag.StringVar(&ConfigFile, "config-file", "", "path to config")
	flag.Var(&logLevel, "log-level", "logging level")
	flag.Var(&logFormat, "log-format", `log format, one of: "text", "json"`)

	// github

	flag.Var(&GithubToken, "github-token", "github token")
	flag.Var(&GithubAppID, "github-app-id", "github app id")
	flag.Var(&GithubAppInstallationID, "github-app-installation-id", "github app installation id")
	flag.Var(&GithubAppPrivateKey, "github-app-private-key", "github app private key")

	// kubernetes

	flag.Var(&kubeConfig, "kube-config", "path to the kubeconfig file")
	flag.Var(&kubeContext, "kube-context", "specify a kubernetes context")
	flag.Var(&Namespace, "namespace", "specify a kubernetes namespace")

	// reconciler

	flag.Var(&SyncInterval, "sync-interval", "seconds between reconciliation attempts (minimum 30s)")

	flag.Parse()
	godotenv.Load()

	loadConfigFromEnv()
	loadConfigFromFile()
	configureLogger()
}

func loadConfigFromEnv() {

	// global
	maybeSetEnv("LOG_LEVEL", &logLevel)
	maybeSetEnv("LOG_FORMAT", &logFormat)

	// github
	maybeSetEnv("GITHUB_TOKEN", &GithubToken)
	maybeSetEnv("GITHUB_APP_ID", &GithubAppID)
	maybeSetEnv("GITHUB_APP_INSTALLATION_ID", &GithubAppInstallationID)
	maybeSetEnv("GITHUB_APP_PRIVATE_KEY", &GithubAppPrivateKey)

	// kubernetes
	maybeSetEnv("KUBE_CONFIG", &kubeConfig)
	maybeSetEnv("KUBE_CONTEXT", &kubeContext)
	maybeSetEnv("NAMESPACE", &Namespace)

	// reconciler
	maybeSetEnv("SYNC_INTERVAL", &SyncInterval)
}

func loadConfigFromFile() {
	filenames := []string{ConfigFile, "./config.yaml"}

	for _, filename := range filenames {
		if len(Runners) > 0 {
			return
		}

		if filename == "" {
			continue
		}

		if _, err := os.Stat(filename); err != nil {
			continue
		}

		raw, err := os.ReadFile(filename)
		if err != nil {
			panic(fmt.Errorf("could not read config file: %s", err))
		}

		var cfg struct {
			// global

			LogLevel  *string `yaml:"log_level"`
			LogFormat *string `yaml:"log_format"`

			// github
			GithubToken             *string `yaml:"github_token"`
			GithubAppID             *string `yaml:"github_app_id"`
			GithubAppInstallationID *string `yaml:"github_app_installation_id"`
			GithubAppPrivateKey     *string `yaml:"github_app_private_key"`

			// kubernetes

			KubeConfig  *string `yaml:"kube_config"`
			KubeContext *string `yaml:"kube_context"`
			Namespace   *string `yaml:"namespace"`

			// reconciler

			SyncInterval *int `yaml:"sync_interval"`

			// dispatcher

			Runners []RunnerConfig `yaml:"runners"`
		}

		if err := yaml.Unmarshal(raw, &cfg); err != nil {
			panic(fmt.Errorf("could not parse config file: %s", err))
		}

		logLevel.MaybeSet(cfg.LogLevel)
		logFormat.MaybeSet(cfg.LogFormat)

		GithubToken.MaybeSet(cfg.GithubToken)
		GithubAppID.MaybeSet(cfg.GithubAppID)
		GithubAppInstallationID.MaybeSet(cfg.GithubAppInstallationID)
		GithubAppPrivateKey.MaybeSet(cfg.GithubAppPrivateKey)

		kubeConfig.MaybeSet(cfg.KubeConfig)
		kubeContext.MaybeSet(cfg.KubeContext)
		Namespace.MaybeSet(cfg.Namespace)

		Runners = cfg.Runners
		for i, runner := range Runners {
			if err := runner.Validate(); err != nil {
				panic(fmt.Errorf("failed to validate runner %d: %s", i, err))
			}
		}
	}
}
