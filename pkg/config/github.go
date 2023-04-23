package config

import "flag"

var (
	GithubToken             StringFlagValue
	GithubAppID             StringFlagValue
	GithubAppInstallationID StringFlagValue
	GithubAppPrivateKey     StringFlagValue
)

func registerGithubFlags() {
	flag.Var(&GithubToken, "github-token", "github token")
	flag.Var(&GithubAppID, "github-app-id", "github app id")
	flag.Var(&GithubAppInstallationID, "github-app-installation-id", "github app installation id")
	flag.Var(&GithubAppPrivateKey, "github-app-private-key", "github app private key")
}
