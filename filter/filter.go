package filter

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type context struct {
	get VariableGetter
}
type VariableGetter func(name string) *Value

var ErrInvalidOperatorOrOperands = errors.New("invalid operator or operands")

func (x *term) eval(ctx context) (*value, error) {
	switch {
	case x.Value != nil:
		return x.Value, nil
	case x.SymbolRef != nil:
		return ctx.get(x.SymbolRef.Symbol).tovalue(x.SymbolRef.Symbol)
	default:
		return x.SubExpression.eval(ctx)
	}
}

func (x *compare) eval(t *term, ctx context) (*value, error) {
	v1, err := t.eval(ctx)
	if err != nil {
		return nil, err
	}
	v2, err := x.Operand.eval(ctx)
	if err != nil {
		return nil, err
	}
	op := x.Operator
	r := false

	switch {
	case v1.Num() != nil && v2.Num() != nil:
		n1, n2 := *v1.Num(), *v2.Num()
		switch op {
		case "!=", "<>":
			r = n1 != n2
		case "<=":
			r = n1 <= n2
		case ">=":
			r = n1 >= n2
		case "=":
			r = n1 == n2
		case "<":
			r = n1 < n2
		case ">":
			r = n1 > n2
		default:
			return nil, ErrInvalidOperatorOrOperands
		}
		return boolValue(r), nil
	case v1.Text != nil && v2.Text != nil:
		t1, t2 := *v1.Text, *v2.Text
		switch op {
		case "!=", "<>":
			r = t1 != t2
		case "<=":
			r = t1 <= t2
		case ">=":
			r = t1 >= t2
		case "=":
			r = t1 == t2
		case "<":
			r = t1 < t2
		case ">":
			r = t1 > t2
		default:
			return nil, ErrInvalidOperatorOrOperands
		}
		return boolValue(r), nil
	case v1.Boolean != nil && v2.Boolean != nil:
		b1, b2 := *v1.Boolean, *v2.Boolean
		switch op {
		case "!=", "<>":
			r = b1 != b2
		case "=":
			r = b1 == b2
		default:
			return nil, ErrInvalidOperatorOrOperands
		}
		return boolValue(r), nil
	}
	return nil, ErrInvalidOperatorOrOperands
}

func (x *between) eval(t *term, ctx context) (*value, error) {
	v1, err := t.eval(ctx)
	if err != nil {
		return nil, err
	}
	v2, err := x.Start.eval(ctx)
	if err != nil {
		return nil, err
	}
	v3, err := x.End.eval(ctx)
	if err != nil {
		return nil, err
	}

	switch {
	case v1.Num() != nil && v2.Num() != nil && v3.Num() != nil:
		n1, n2, n3 := *v1.Num(), *v2.Num(), *v3.Num()
		return boolValue(n1 >= n2 && n1 <= n3), nil
	case v1.Text != nil && v2.Text != nil && v3.Text != nil:
		t1, t2, t3 := *v1.Text, *v2.Text, *v3.Text
		return boolValue(t1 >= t2 && t1 <= t3), nil
	}
	return nil, ErrInvalidOperatorOrOperands
}

func (x *in) eval(t *term, ctx context) (*value, error) {
	v1, err := t.eval(ctx)
	if err != nil {
		return nil, err
	}
	for _, o := range x.Expressions {
		if v2, err := o.eval(ctx); err != nil {
			return nil, err
		} else {
			switch {
			case v1.Num() != nil && v2.Num() != nil:
				n1, n2 := *v1.Num(), *v2.Num()
				if n1 == n2 {
					return boolValue(true), nil
				}
			case v1.Text != nil && v2.Text != nil:
				t1, t2 := *v1.Text, *v2.Text
				if t1 == t2 {
					return boolValue(true), nil
				}
			case v1.Boolean != nil && v2.Boolean != nil:
				b1, b2 := v1.Bool(), v2.Bool()
				if b1 == b2 {
					return boolValue(true), nil
				}
			default:
				fmt.Println(v1, v2)
				return nil, ErrInvalidOperatorOrOperands
			}
		}
	}
	return boolValue(false), nil
}

func not(value *value, err error) (*value, error) {
	return boolValue(!value.Bool()), err
}

func (x *conditionRHS) eval(t *term, ctx context) (*value, error) {
	r := false
	switch {
	case x.Compare != nil:
		return x.Compare.eval(t, ctx)
	case x.Between != nil:
		if x.Not {
			return not(x.Between.eval(t, ctx))
		} else {
			return x.Between.eval(t, ctx)
		}
	case x.In != nil:
		if x.Not {
			return not(x.In.eval(t, ctx))
		} else {
			return x.In.eval(t, ctx)
		}
	}

	// *like
	// assume regex is static
	if x.likeCache == nil {
		switch {
		case x.Like != nil:
			v2, err := x.Like.eval(ctx)
			if err != nil {
				return nil, err
			}
			v2r := strings.ReplaceAll(regexp.QuoteMeta(v2.String()), "%", ".*")
			v2r = "^" + v2r + "$"
			x.likeCache = regexp.MustCompile(v2r)
		case x.Ilike != nil:
			v2, err := x.Ilike.eval(ctx)
			if err != nil {
				return nil, err
			}
			v2r := strings.ReplaceAll(regexp.QuoteMeta(v2.String()), "%", ".*")
			v2r = "(?i)^" + v2r + "$"
			x.likeCache = regexp.MustCompile(v2r)
		case x.Rlike != nil:
			v2, err := x.Rlike.eval(ctx)
			if err != nil {
				return nil, err
			}
			x.likeCache = regexp.MustCompile(v2.String())
		}
	}

	v1, err := t.eval(ctx)
	if err != nil {
		return nil, err
	}
	r = x.likeCache.MatchString(v1.String())
	if x.Not {
		r = !r
	}
	return boolValue(r), nil
}

func (x *conditionOperand) eval(ctx context) (*value, error) {
	if x.ConditionRHS != nil {
		return x.ConditionRHS.eval(x.Operand, ctx)
	} else {
		return x.Operand.eval(ctx)
	}
}

func (x *condition) eval(ctx context) (*value, error) {
	switch {
	case x.Operand != nil:
		return x.Operand.eval(ctx)
	default:
		if v, err := x.Not.eval(ctx); err != nil {
			return nil, err
		} else {
			return boolValue(!v.Bool()), nil
		}
	}
}

func (x *andCondition) eval(ctx context) (*value, error) {
	r := true
	for _, o := range x.And {
		if v, err := o.eval(ctx); err != nil {
			return nil, err
		} else {
			r = r && v.Bool()
		}
	}
	return boolValue(r), nil
}

func (x *expression) eval(ctx context) (*value, error) {
	r := false
	for _, o := range x.Or {
		if v, err := o.eval(ctx); err != nil {
			return nil, err
		} else {
			r = r || v.Bool()
		}
	}
	return boolValue(r), nil
}

type FilterExpression struct {
	expression *expression
}

func (x *FilterExpression) Test(getter VariableGetter) (bool, error) {
	ctx := context{get: getter}
	if r, err := x.expression.eval(ctx); err != nil {
		return false, err
	} else {
		return r.Bool(), nil
	}
}

func CreateFilter(filter string) (*FilterExpression, error) {
	if expr, err := parser.ParseString("", filter); err != nil {
		return nil, err
	} else {
		return &FilterExpression{expression: expr}, nil
	}
}
