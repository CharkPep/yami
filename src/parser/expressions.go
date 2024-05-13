package parser

import (
	"bytes"
	"fmt"
	"github.com/charkpep/yad/src/lexer"
	"strings"
)

type Expression interface {
	Node
	expression()
}

type ExpressionStatement struct {
	Expr Expression
	Tok  lexer.Token // first Tok of the expression
}

func (exr ExpressionStatement) Token() lexer.Token {
	return exr.Tok
}

func (exr ExpressionStatement) statement() {}

func (exr ExpressionStatement) String() string {
	return exr.Expr.String()
}

type IntegerExpression struct {
	token lexer.Token
	Val   int64
}

func (i IntegerExpression) Token() lexer.Token {
	return i.token
}

func (i IntegerExpression) expression() {}

func (i IntegerExpression) String() string {
	return fmt.Sprintf("%d", i.Val)
}

type IdentifierExpression struct {
	Identifier lexer.Token
}

func (i IdentifierExpression) Token() lexer.Token {
	return i.Identifier
}

func (i IdentifierExpression) expression() {}

func (i IdentifierExpression) String() string {
	return i.Identifier.Literal
}

type InfixExpression struct {
	Left     Expression
	Operator lexer.Token
	Right    Expression
}

func (inf *InfixExpression) Token() lexer.Token {
	return inf.Operator
}

func (inf *InfixExpression) expression() {}

func (inf *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", inf.Left.String(), inf.Operator.Literal, inf.Right.String())
}

type PrefixExpression struct {
	Prefix lexer.Token
	Expr   Expression
}

func (p PrefixExpression) Token() lexer.Token {
	return p.Prefix
}

func (p PrefixExpression) expression() {}

func (p PrefixExpression) String() string {
	return fmt.Sprintf("%s(%s)", p.Prefix.Literal, p.Expr)
}

// IfExpression TODO remove in flavor of ternary Operator, if should be a statement
type IfExpression struct {
	token       lexer.Token
	Condition   Expression
	Consequence BlockStatement
	Alternative *BlockStatement
}

func (i IfExpression) Token() lexer.Token {
	return i.token
}

func (i IfExpression) expression() {}

func (i IfExpression) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString("if ")
	buff.WriteString(i.Condition.String())
	buff.WriteString(" { ")
	buff.WriteString(i.Consequence.String())
	buff.WriteString(" }")
	if i.Alternative != nil {
		buff.WriteString(" else {")
		buff.WriteString(i.Alternative.String())
		buff.WriteString("}")
	}

	buff.WriteString("\n")
	return buff.String()
}

type FuncExpression struct {
	token lexer.Token
	Args  []IdentifierExpression
	Body  BlockStatement
}

func (f FuncExpression) Token() lexer.Token {
	return f.token
}

func (f FuncExpression) expression() {}

func (f FuncExpression) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString("fn ")
	buff.WriteString("(")
	for i, arg := range f.Args {
		buff.WriteString(arg.String())
		if i+1 != len(f.Args) {
			buff.WriteString(",")
		}
	}

	buff.WriteString(") ")
	buff.WriteString(f.Body.String())
	return buff.String()
}

type CallExpression struct {
	token    lexer.Token
	Call     Expression
	CallArgs []Expression
}

func (c CallExpression) expression() {}

func (c CallExpression) Token() lexer.Token {
	return c.token
}

func (c CallExpression) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(c.Call.String())
	buff.WriteString("(")
	for i, arg := range c.CallArgs {
		buff.WriteString(arg.String())
		if i+1 != len(c.CallArgs) {
			buff.WriteString(",")
		}
	}
	buff.WriteString(")")
	return buff.String()
}

type AssignExpression struct {
	token      lexer.Token
	Identifier Expression
	Val        Expression
}

func (ass AssignExpression) Token() lexer.Token {
	return ass.token
}

func (ass AssignExpression) expression() {}

func (ass AssignExpression) String() string {
	return fmt.Sprintf("%s=%s", ass.Identifier, ass.Val)
}

type BoolExpression struct {
	token lexer.Token
	Val   bool
}

func (bl BoolExpression) Token() lexer.Token {
	return bl.token
}

func (bl BoolExpression) String() string {
	return fmt.Sprint(bl.Val)
}

func (bl BoolExpression) expression() {}

type StringExpression struct {
	tok lexer.Token
	Val string
}

func (str StringExpression) Token() lexer.Token {
	return str.tok
}

func (str StringExpression) expression() {}

func (str StringExpression) String() string {
	return fmt.Sprintf(`"%s"`, str.Val)
}

// a[1] -> array index, str[1] -> string index, hmap[Any Expression] -> hashmap index
type IndexExpression struct {
	token lexer.Token
	Of    Expression
	Idx   Expression
}

func (idx IndexExpression) Token() lexer.Token {
	return idx.token
}

func (idx IndexExpression) expression() {}

func (idx IndexExpression) String() string {
	var buff bytes.Buffer
	buff.WriteString(idx.Of.String())
	buff.WriteString("[")
	buff.WriteString(idx.Idx.String())
	buff.WriteString("]")
	return buff.String()
}

type ArrayExpression struct {
	Arr   []Expression
	token lexer.Token
}

func (arr ArrayExpression) Token() lexer.Token {
	return arr.token
}

func (arr ArrayExpression) expression() {}

func (arr ArrayExpression) String() string {
	var buff bytes.Buffer
	buff.WriteString("[")
	expressions := make([]string, 0, len(arr.Arr))
	for _, expr := range arr.Arr {
		expressions = append(expressions, expr.String())
	}
	buff.WriteString(strings.Join(expressions, ","))
	buff.WriteString("]")
	return buff.String()
}

type HashMapExpression struct {
	Map   map[Expression]Expression
	token lexer.Token
}

func (mp HashMapExpression) String() string {
	var buff bytes.Buffer
	buff.WriteString("{")
	elements := make([]string, 0, len(mp.Map))
	for k, v := range mp.Map {
		elements = append(elements, fmt.Sprintf("%s:%s", k.String(), v.String()))
	}

	buff.WriteString(strings.Join(elements, ","))
	buff.WriteString("}")
	return buff.String()
}

func (mp HashMapExpression) expression() {}

func (mp HashMapExpression) Token() lexer.Token {
	return mp.token
}
