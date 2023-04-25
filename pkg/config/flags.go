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

func (v StringFlagValue) Value() string {
	return v.value
}

func (v StringFlagValue) String() string {
	return v.value
}

func (v *StringFlagValue) MaybeSet(value *string) error {
	if v.set {
		return nil
	}

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

type Int64FlagValue struct {
	set          bool
	value        int64
	defaultValue int64
	validate     func(int64) error
}

func (v Int64FlagValue) Value() int64 {
	return v.value
}

func (v Int64FlagValue) String() string {
	return strconv.FormatInt(v.value, 10)
}

func (v *Int64FlagValue) MaybeSet(value *string) error {
	if v.set {
		return nil
	}

	if value != nil {
		return v.Set(*value)
	}

	return v.Set(strconv.FormatInt(v.defaultValue, 10))
}

func (v *Int64FlagValue) Set(raw string) error {
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
