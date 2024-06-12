package filter

import (
	"errors"
	"fmt"
)

// Value is a type that can represent a number, a string, or a boolean value.
type Value struct {
	Number  *int64
	Text    *string
	Boolean *bool
}

// String returns a string representation of the value. If the value is nil, an empty
// string is returned.
func (v *Value) String() string {
	if v == nil {
		return ""
	}
	switch {
	case v.Number != nil:
		return fmt.Sprintf("%d", *v.Number)
	case v.Text != nil:
		return *v.Text
	case v.Boolean != nil:
		return fmt.Sprintf("%t", *v.Boolean)
	default:
		return ""
	}
}

func (v *Value) tovalue(name string) (*value, error) {
	if v != nil {
		return &value{
			Number:  v.Number,
			Text:    v.Text,
			Boolean: (*boolean)(v.Boolean),
		}, nil
	} else {
		return &value{}, errors.New(fmt.Sprintf("\"%s\" is unknown", name))
	}
}

// NumberValue creates a new Value instance that represents the given number.
func NumberValue(f int64) *Value {
	return &Value{
		Number: &f,
	}
}

// TextValue creates a new Value instance that represents the given string.
func TextValue(s string) *Value {
	return &Value{
		Text: &s,
	}
}

// BoolValue creates a new Value instance that represents the given boolean value.
func BoolValue(v bool) *Value {
	b := bool(v)
	return &Value{
		Boolean: &b,
	}
}
