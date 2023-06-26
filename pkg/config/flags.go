package config

import (
	"strconv"
)

type stringFlag struct {
	set          bool
	value        string
	defaultValue string
	validate     func(string) error
}

func (v stringFlag) Value() string {
	return v.value
}

func (v stringFlag) String() string {
	return v.value
}

func (v *stringFlag) MaybeSet(value *string) error {
	if v.set {
		return nil
	}

	if value != nil {
		return v.Set(*value)
	}

	return v.Set(v.defaultValue)
}

func (v *stringFlag) Set(value string) error {
	if v.validate != nil {
		if err := v.validate(value); err != nil {
			return err
		}
	}

	v.set = true
	v.value = value
	return nil
}

type IntFlagValue struct {
	set          bool
	value        int
	defaultValue int
	validate     func(int) error
}

func (v IntFlagValue) Value() int {
	return v.value
}

func (v IntFlagValue) String() string {
	return strconv.Itoa(v.value)
}

func (v *IntFlagValue) MaybeSet(value *string) error {
	if v.set {
		return nil
	}

	if value != nil {
		return v.Set(*value)
	}

	return v.Set(strconv.Itoa(v.defaultValue))
}

func (v *IntFlagValue) Set(raw string) error {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}

	if v.validate != nil {
		if err := v.validate(value); err != nil {
			return err
		}
	}

	v.set = true
	v.value = value
	return nil
}

type int64Flag struct {
	set          bool
	value        int64
	defaultValue int64
	validate     func(int64) error
}

func (v int64Flag) Value() int64 {
	return v.value
}

func (v int64Flag) String() string {
	return strconv.FormatInt(v.value, 10)
}

func (v *int64Flag) MaybeSet(value *string) error {
	if v.set {
		return nil
	}

	if value != nil {
		return v.Set(*value)
	}

	return v.Set(strconv.FormatInt(v.defaultValue, 10))
}

func (v *int64Flag) Set(raw string) error {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}

	if v.validate != nil {
		if err := v.validate(value); err != nil {
			return err
		}
	}

	v.set = true
	v.value = value
	return nil
}

type boolFlag struct {
	set          bool
	value        bool
	defaultValue bool
	validate     func(bool) error
}

func (v boolFlag) Value() bool {
	return v.value
}

func (v boolFlag) String() string {
	return strconv.FormatBool(v.value)
}

func (v *boolFlag) MaybeSet(value *string) error {
	if v.set {
		return nil
	}

	if value != nil {
		return v.Set(*value)
	}

	return v.Set(strconv.FormatBool(v.defaultValue))
}

func (v *boolFlag) Set(raw string) error {
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return err
	}

	if v.validate != nil {
		if err := v.validate(value); err != nil {
			return err
		}
	}

	v.set = true
	v.value = value
	return nil
}
