package utils

import (
	"fmt"
	"strings"
)

// StringSliceFlag is a custom flag type that allows for multiple string values
type StringSliceFlag []string

// String returns a string representation of the flag values
func (s *StringSliceFlag) String() string {
	return strings.Join(*s, ", ")
}

// Set adds a value to the StringSliceFlag
func (s *StringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// Type returns the type of the flag as a string
func (s *StringSliceFlag) Type() string {
	return "stringSlice"
}

// IntRangeFlag is a custom flag type for specifying a range of integers
type IntFlag struct {
	Value int
}

// String returns a string representation of the IntFlag
func (i *IntFlag) String() string {
	return fmt.Sprintf("%d", i.Value)
}

// Set parses a string in the format "min-max" and sets the Min and Max values
func (i *IntFlag) Set(value string) error {
	_, err := fmt.Sscanf(value, "%d", &i.Value)
	if err != nil {
		return fmt.Errorf("invalid integer format: %w", err)
	}
	return nil
}

// Type returns the type of the flag as a string
func (i *IntFlag) Type() string {
	return "int"
}

// ParseFlags is a utility function to parse command-line flags
func ParseFlags() error {
	return nil
}
