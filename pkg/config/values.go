package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

type valueWithDefault interface {
	Default() string
}

var (
	_ flag.Value = (*logLevelValue)(nil)
)

type logLevelValue zerolog.Level

func (v *logLevelValue) Default() string {
	return fmt.Sprint(zerolog.InfoLevel)
}

func (v *logLevelValue) Set(s string) error {
	level, err := zerolog.ParseLevel(s)
	if err != nil {
		return err
	}

	*v = logLevelValue(level)
	return nil
}

func (v *logLevelValue) String() string {
	if v == nil {
		return v.Default()
	}

	return fmt.Sprint(*v)
}

type logFormatValue string

const (
	jsonLogFormat logFormatValue = "json"
	textLogFormat logFormatValue = "text"
)

func (v *logFormatValue) Default() string {
	return string(jsonLogFormat)
}

func (v *logFormatValue) Set(s string) error {
	val := logFormatValue(s)
	if err := val.Validate(); err != nil {
		return err
	}

	*v = logFormatValue(val)
	return nil
}

func (v *logFormatValue) String() string {
	if v == nil {
		return v.Default()
	}

	return string(*v)
}

func (t logFormatValue) Values() []string {
	return []string{
		string(jsonLogFormat),
		string(textLogFormat),
	}
}

func (t logFormatValue) Validate() error {
	for _, v := range t.Values() {
		if t == logFormatValue(v) {
			return nil
		}
	}

	return fmt.Errorf("format must be one of [%s], got %s", strings.Join(t.Values(), ", "), t)
}
