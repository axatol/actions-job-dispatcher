package config

import (
	"flag"
	"fmt"
	"strconv"
)

var (
	Interval IntervalValue
)

func registerReconcilerFlags() {
	flag.Var(&Interval, "interval", "seconds between reconciliation attempts (minimum 30s)")
}

type IntervalValue struct {
	set   bool
	value int
}

func (v IntervalValue) DefaultString() string {
	return "30"
}

func (v IntervalValue) String() string {
	return fmt.Sprint(v.value)
}

func (v *IntervalValue) MaybeSet(value *string) error {
	if value != nil {
		return v.Set(*value)
	}

	return v.Set(v.DefaultString())
}

func (v *IntervalValue) Set(raw string) error {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}

	if value < 30 {
		return fmt.Errorf("interval cannot be less than 30, got: %d", value)
	}

	v.set = true
	v.value = value
	return nil
}
