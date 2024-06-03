package filter

import (
	"errors"
	"fmt"
)

type Value struct {
	Number  *int64
	Text    *string
	Boolean *bool
}

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

func NumberValue(f int64) *Value {
	return &Value{
		Number: &f,
	}
}

func TextValue(s string) *Value {
	return &Value{
		Text: &s,
	}
}

func BoolValue(v bool) *Value {
	b := bool(v)
	return &Value{
		Boolean: &b,
	}
}
