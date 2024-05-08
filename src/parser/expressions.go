package parser

import (
	"bytes"
	"fmt"
	"github.com/charkpep/yad/src/lexer"
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
	identifier lexer.Token
}

func (i IdentifierExpression) Token() lexer.Token {
	return i.identifier
}

func (i IdentifierExpression) expression() {}

func (i IdentifierExpression) String() string {
	return i.identifier.Literal
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
	condition   Expression
	consequence BlockStatement
	alternative *BlockStatement
}

func (i IfExpression) Token() lexer.Token {
	return i.token
}

func (i IfExpression) expression() {}

func (i IfExpression) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString("if ")
	buff.WriteString(i.condition.String())
	buff.WriteString(" { ")
	buff.WriteString(i.consequence.String())
	buff.WriteString(" }")
	if i.alternative != nil {
		buff.WriteString(" else {")
		buff.WriteString(i.alternative.String())
		buff.WriteString("}")
	}

	buff.WriteString("\n")
	return buff.String()
}

type FuncExpression struct {
	token lexer.Token
	args  []IdentifierExpression
	body  BlockStatement
}

func (f FuncExpression) Token() lexer.Token {
	return f.token
}

func (f FuncExpression) expression() {}

func (f FuncExpression) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString("fn ")
	buff.WriteString("(")
	for i, arg := range f.args {
		buff.WriteString(arg.String())
		if i+1 != len(f.args) {
			buff.WriteString(",")
		}
	}

	buff.WriteString(") ")
	buff.WriteString(f.body.String())
	return buff.String()
}

type CallExpression struct {
	token    lexer.Token
	f        Expression
	callArgs []Expression
}

func (c CallExpression) expression() {}

func (c CallExpression) Token() lexer.Token {
	return c.token
}

func (c CallExpression) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString(c.f.String())
	buff.WriteString("(")
	for i, arg := range c.callArgs {
		buff.WriteString(arg.String())
		if i+1 != len(c.callArgs) {
			buff.WriteString(",")
		}
	}
	buff.WriteString(")")
	return buff.String()
}

type AssignExpression struct {
	token      lexer.Token
	Identifier IdentifierExpression
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
