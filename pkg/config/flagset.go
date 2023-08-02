package config

import (
	"flag"
	"os"
	"strings"
)

type flagSet struct{ *flag.FlagSet }

// returns all flags not yet set
func (fs flagSet) Unset() []*flag.Flag {
	var flags []*flag.Flag

	// collect all flags
	flag.VisitAll(func(f *flag.Flag) {
		flags = append(flags, f)
	})

	// remove flags with values
	flag.Visit(func(f *flag.Flag) {
		for i, set := range flags {
			if set.Name == f.Name {
				flags = append(flags[:i], flags[i+1:]...)
			}
		}
	})

	return flags
}

// set unset from env or default
func (fs flagSet) LoadUnsetFromEnv() {
	for _, f := range fs.Unset() {
		envKey := strings.ToUpper(f.Name)
		envKey = strings.ReplaceAll(envKey, "-", "_")

		if val, ok := os.LookupEnv(envKey); ok {
			f.Value.Set(val)
			continue
		}

		// last chance using default if any
		if val, ok := (f.Value).(valueWithDefault); ok {
			f.Value.Set(val.Default())
		}
	}
}
