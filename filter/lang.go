package filter

import (
	"fmt"
	"regexp"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type boolean bool

type expression struct {
	Or []*andCondition `@@ ( "OR" @@ )*`
}

type andCondition struct {
	And []*condition `@@ ( "AND" @@ )*`
}

type condition struct {
	Operand *conditionOperand `  @@`
	Not     *condition        `| "NOT" @@`
}

type conditionOperand struct {
	Operand      *term         `@@`
	ConditionRHS *conditionRHS `@@?`
}

type conditionRHS struct {
	Compare   *compare `  @@`
	Not       bool     `| [ @"NOT" ] (`
	Between   *between `      "BETWEEN" @@`
	In        *in      `    | "IN" "(" @@ ")"`
	Ilike     *term    `    | "ILIKE" @@`
	Rlike     *term    `    | "RLIKE" @@`
	Like      *term    `    | "LIKE" @@ )`
	likeCache *regexp.Regexp
}

type compare struct {
	Operator string `@( "<>" | "<=" | ">=" | "=" | "<" | ">" | "!=" )`
	Operand  *term  `@@`
}

type between struct {
	Start *term `@@`
	End   *term `"AND" @@`
}

type in struct {
	Expressions []*term `@@ ( "," @@ )*`
}

type term struct {
	Value         *value      `  @@`
	SymbolRef     *symbolRef  `| @@`
	SubExpression *expression `| "(" @@ ")"`
}

type symbolRef struct {
	Symbol string `@Ident`
}

type size int64

func (s *size) Capture(v []string) error { n, err := ParseSize(v[0]); *s = size(n); return err }

type value struct {
	Size    *size    ` ( @Size`
	Number  *int64   ` | @Number`
	Text    *string  ` | @Text`
	Boolean *boolean ` | @("TRUE" | "FALSE") )`
}

func boolValue(v bool) *value {
	b := boolean(v)
	return &value{
		Boolean: &b,
	}
}

func (v value) Num() *int64 {
	if v.Size != nil {
		return (*int64)(v.Size)
	}
	return v.Number
}

func (v value) String() string {
	switch {
	case v.Num() != nil:
		return fmt.Sprintf("%d", *v.Num())
	case v.Text != nil:
		return *v.Text
	case v.Boolean != nil:
		return fmt.Sprintf("%t", *v.Boolean)
	default:
		return "(empty)"
	}
}

func (x *value) Bool() bool {
	switch {
	case x.Num() != nil:
		return *x.Num() != 0
	case x.Text != nil:
		return *x.Text != ""
	case x.Boolean != nil:
		return bool(*x.Boolean)
	default:
		return false
	}
}

var (
	exprLexer = lexer.MustSimple([]lexer.SimpleRule{
		{`Keyword`, `(?i)\b(TRUE|FALSE|NOT|BETWEEN|AND|OR|LIKE|ILIKE|RLIKE|IN)\b`},
		{`Ident`, `[a-zA-Z_][a-zA-Z0-9_]*`},
		{`Size`, `\d*\.?\d+[BKMGT]`},
		{`Number`, `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
		{`Text`, `'[^']*'|"[^"]*"`},
		{`Operators`, `<>|!=|<=|>=|[,.()=<>]`},
		{"whitespace", `\s+`},
	})
	parser = participle.MustBuild[expression](
		participle.Lexer(exprLexer),
		participle.Unquote("Text"),
		participle.CaseInsensitive("Keyword"),
	)
)
