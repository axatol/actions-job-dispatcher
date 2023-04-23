package config

import (
	"flag"
	"fmt"

	"github.com/rs/zerolog"
)

var (
	ConfigFile string
	LogLevel   LogLevelValue
	LogFormat  LogFormatValue
)

func registerGlobalFlags() {
	flag.StringVar(&ConfigFile, "config-file", "", "path to config")
	flag.Var(&LogLevel, "log-level", "logging level")
	flag.Var(&LogFormat, "log-format", `log format, one of: "text", "json"`)
}

const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

type LogFormatValue struct {
	set   bool
	value string
}

func (v LogFormatValue) DefaultString() string {
	return LogFormatJSON
}

func (v LogFormatValue) String() string {
	return v.value
}

func (v *LogFormatValue) MaybeSet(value *string) error {
	if value != nil {
		return v.Set(*value)
	}

	return v.Set(v.DefaultString())
}

func (v *LogFormatValue) Set(value string) error {
	if value != LogFormatJSON && value != LogFormatText {
		return fmt.Errorf(`value must be one of: "text", "json"`)
	}

	v.set = true
	v.value = value
	return nil
}

type LogLevelValue struct {
	set   bool
	value zerolog.Level
}

func (v LogLevelValue) DefaultString() string {
	return zerolog.InfoLevel.String()
}

func (v LogLevelValue) String() string {
	return v.value.String()
}

func (v *LogLevelValue) MaybeSet(value *string) error {
	if value != nil {
		return v.Set(*value)
	}

	return v.Set(v.DefaultString())
}

func (v *LogLevelValue) Set(value string) error {
	level, err := zerolog.ParseLevel(value)
	if err != nil {
		return err
	}

	v.set = true
	v.value = level
	return nil
}
