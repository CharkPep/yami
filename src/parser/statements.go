package parser

import (
	"bytes"
	"fmt"
	"github.com/charkpep/yami/src/lexer"
)

type Statement interface {
	Node
	statement()
}

type LetStatement struct {
	Literal    lexer.Token
	Identifier IdentifierExpression
	Expression Expression
}

type ReturnStatement struct {
	token      lexer.Token
	ReturnExpr Expression
}

func (r ReturnStatement) statement() {}

func (r ReturnStatement) Token() lexer.Token {
	return r.token
}

func (r ReturnStatement) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString("return ")
	buff.WriteString(r.ReturnExpr.String())
	buff.WriteString(";")
	return buff.String()
}

func (l LetStatement) Token() lexer.Token {
	return l.Literal
}

func (LetStatement) statement() {}

func (l LetStatement) String() string {
	return fmt.Sprintf("let %s=%s;", l.Identifier.String(), l.Expression.String())
}

type BlockStatement struct {
	token      lexer.Token
	Statements []Statement
}

func (b BlockStatement) Token() lexer.Token {
	return b.token
}

func (b BlockStatement) statement() {}

func (b BlockStatement) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteString("{\n")
	for _, st := range b.Statements {
		buff.WriteString(st.String())
	}
	buff.WriteString("}")
	return buff.String()
}
