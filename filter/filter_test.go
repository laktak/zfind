package filter

import (
	"fmt"
	"testing"
)

type example struct {
	expected bool
	errmsg   string
	w        string
}

var examples = []example{
	{true, "", "name like \"foo%\""},
	{true, "", "name like \"%oo%\""},
	{true, "", "name like \"%bar\""},
	{false, "", "name like \"oo%\""},
	{false, "", "name like \"Foo%\""},
	{true, "", "name ilike \"Foo%\""},
	{false, "", "name ilike \"oo%\""},
	{true, "", "name rlike \"^foo.*$\""},
	{true, "", "name rlike \"foo\""},
	{false, "", "name not like \"foo%\""},
	{true, "", "name not like \"test%\""},
	{true, "", "x=3"},
	{true, "", "x=3 and y=40000"},
	{false, "", "x like \"x\""},
	{false, "", "x=5"},
	{true, "", "x in (3, 5)"},
	{false, "", "x not in (3, 5)"},
	{false, "", "x in (4, 6, 5)"},
	{true, "", "x not in (4, 6, 5)"},
	{true, "", "x between 3 and 5"},
	{false, "", "x not between 3 and 5"},
	{false, "", "x between 4 and 5"},
	{true, "", "x=3 and y<70K"},
	{true, "", "x=5 and y=40000 or name=\"foobar\""},
	{false, "", "x=5 and (y=40000 or name=\"foobar\")"},
	{false, "\"noname\" is unknown", "noname like \"hug%\""},
	{true, "\"i\" is unknown", "x=5 and (i=7 or foo='foo')"},
	{false, "invalid operator or operands", "x=\"x\""},
}

func check(t *testing.T, w string, expect bool, errmsg string) string {
	test := func(name string) *Value {
		switch name {
		case "x":
			return NumberValue(3)
		case "y":
			return NumberValue(40000)
		case "name":
			return TextValue("foobar")
		default:
			return nil
		}
	}
	chkerr := func(err error) string {
		serr := fmt.Sprintf("%v", err)
		if serr != errmsg {
			return fmt.Sprintf("Err %s vs %s", serr, errmsg)
		} else {
			return ""
		}
	}

	if filter, err := CreateFilter(w); err == nil {
		if r, err := filter.Test(test); err == nil {
			if errmsg != "" {
				return "missing error: " + errmsg
			}
			if r != expect {
				return fmt.Sprintf("result=%t expected=%t", r, expect)
			}
			return ""
		} else {
			return chkerr(err)
		}
	} else {
		return chkerr(err)
	}
}

func TestFilters(t *testing.T) {
	for _, ex := range examples {
		r := check(t, ex.w, ex.expected, ex.errmsg)
		if r != "" {
			t.Error(ex.w + ": " + r)
		}
	}
}
