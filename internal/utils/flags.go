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
type IntRangeFlag struct {
	Min, Max int
}

// String returns a string representation of the IntRangeFlag
func (i *IntRangeFlag) String() string {
	return fmt.Sprintf("%d-%d", i.Min, i.Max)
}

// Set parses a string in the format "min-max" and sets the Min and Max values
func (i *IntRangeFlag) Set(value string) error {
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid range format, expected min-max")
	}

	_, err := fmt.Sscanf(value, "%d-%d", &i.Min, &i.Max)
	if err != nil {
		return fmt.Errorf("invalid range format: %w", err)
	}

	if i.Min > i.Max {
		return fmt.Errorf("min value cannot be greater than max value")
	}

	return nil
}

// Type returns the type of the flag as a string
func (i *IntRangeFlag) Type() string {
	return "intRange"
}

// ParseFlags is a utility function to parse command-line flags
func ParseFlags() error {
	// This function can be implemented if you need custom flag parsing logic
	// For now, we'll leave it as a placeholder
	return nil
}
