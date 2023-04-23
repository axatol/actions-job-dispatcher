package config

import (
	"strconv"
)

type StringFlagValue struct {
	set          bool
	value        string
	defaultValue string
	validate     func(string) error
}

func (v StringFlagValue) String() string {
	return v.value
}

func (v *StringFlagValue) MaybeSet(value *string) error {
	if value != nil {
		return v.Set(*value)
	}

	return v.Set(v.defaultValue)
}

func (v *StringFlagValue) Set(value string) error {
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

func (v IntFlagValue) String() string {
	return strconv.Itoa(v.value)
}

func (v *IntFlagValue) MaybeSet(value *string) error {
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
