package config

import "os"

type settable interface {
	MaybeSet(*string) error
	Set(string) error
}

func maybeSetEnv(key string, value settable) {
	if env, ok := os.LookupEnv(key); ok {
		value.Set(env)
	}
}
