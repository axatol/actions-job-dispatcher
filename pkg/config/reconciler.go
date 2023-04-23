package config

import (
	"flag"
	"fmt"
)

var (
	SyncInterval = IntFlagValue{defaultValue: 30, validate: validateSyncInterval}
)

func registerReconcilerFlags() {
	flag.Var(&SyncInterval, "sync-interval", "seconds between reconciliation attempts (minimum 30s)")
}

func validateSyncInterval(i int) error {
	if i < 30 {
		return fmt.Errorf("sync interval cannot be less than 30, got: %d", i)
	}

	return nil
}
