package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/charkpep/yad/src/lexer"
	"io"
)

const (
	LOWEST = iota
	ASSIGN
	OR // ==, !=, &&, ||
	AND
	RELATIONAL // < >, <= or >=, ==, !=
	BIN_AND
	XOR
	BIN_OR
	BIN_SHIFT
	ADDITION // +
	MULTIPLICATION
	PREFIX // -5 !5
	CALL   // func()
)

type (
	prefixParseFn func() (Expression, error)
	infixParseFn  func(expression Expression) (Expression, error)
)

type Node interface {
	Token() lexer.Token
	String() string
}

type RootNode struct {
	Statements []Statement
}

func (r *RootNode) Token() lexer.Token {
	return lexer.Token{}
}

func (r *RootNode) String() string {
	buff := bytes.NewBuffer(make([]byte, 0))
	for _, st := range r.Statements {
		buff.WriteString(st.String())
		buff.WriteString("\n")
	}

	return buff.String()
}

type ParsingError struct {
	msg   string
	token lexer.Token
}

func (p ParsingError) Error() string {
	return fmt.Sprintf("Parsing error | line: %d, column: %d | message: %s | token: %s\n", p.token.Line, p.token.Column, p.msg, p.token.Literal)
}

func NewParsingError(msg string, token lexer.Token) ParsingError {
	return ParsingError{
		msg:   msg,
		token: token,
	}
}

func NewParser(r io.Reader) *Parser {
	lex := lexer.New(bufio.NewReader(r))
	return NewParserFromLexer(lex)
}

type Parser struct {
	lex lexer.ILexer
	//exprParser *ExpressionParser

	prefixParseFn map[lexer.TokenType]prefixParseFn
	infixParseFn  map[lexer.TokenType]infixParseFn

	Errors []ParsingError
}

func (p *Parser) registerPrefixFunc(token lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFn[token] = fn
}

func (p *Parser) registerInfixFunc(token lexer.TokenType, fn infixParseFn) {
	p.infixParseFn[token] = fn
}

func (p *Parser) registerInfixesFunc(fn infixParseFn, tokens ...lexer.TokenType) {
	for _, t := range tokens {
		p.infixParseFn[t] = fn
	}
}

func (p *Parser) SetLexer(lex lexer.ILexer) *Parser {
	p.lex = lex
	return p
}

func NewParserFromLexer(lex lexer.ILexer) *Parser {
	p := &Parser{
		lex:           lex,
		prefixParseFn: make(map[lexer.TokenType]prefixParseFn),
		infixParseFn:  make(map[lexer.TokenType]infixParseFn),
	}

	p.registerPrefixFunc(lexer.IDENT, p.parseIdentifier)
	p.registerPrefixFunc(lexer.BLEFT, p.ParseGroupedExpression)
	p.registerPrefixFunc(lexer.NUMBER, p.ParseNumber)
	p.registerPrefixFunc(lexer.BANG, p.ParsePrefix)
	p.registerPrefixFunc(lexer.HYPHEN, p.ParsePrefix)
	p.registerPrefixFunc(lexer.IF, p.parseIfExpression)
	p.registerPrefixFunc(lexer.FUNC, p.ParseFuncExpression)
	p.registerInfixFunc(lexer.BLEFT, p.parseCallExpression)
	p.registerInfixFunc(lexer.ASSIGN, p.parseAssignExpression)
	p.registerPrefixFunc(lexer.TRUE, p.parseBoolExpression)
	p.registerPrefixFunc(lexer.FALSE, p.parseBoolExpression)

	p.registerInfixesFunc(p.ParseInfix, lexer.PLUS, lexer.HYPHEN, lexer.SLASH, lexer.ASTERISK, lexer.EQ, lexer.NEQ,
		lexer.DVERTLINE, lexer.DAMPERSAND, lexer.GT, lexer.GTE, lexer.LT, lexer.LTE)

	return p
}

func (p *Parser) Parse() (Node, error) {
	root := RootNode{}
	p.Errors = nil
	var err error
	if p.lex.CurToken().Token == lexer.NIL {
		p.lex.Advance()
	}

	for token := p.lex.CurToken(); token.Token != lexer.EOF && token.Token != lexer.ILLEGAL; token, err = p.lex.Advance() {
		if err != nil {
			return nil, err
		}
		st := p.parseStatement()
		// parsing error occurred
		if st == nil || token.Token == lexer.NIL {
			p.lex.Advance()
			continue
		}
		root.Statements = append(root.Statements, st)
	}

	return &root, nil
}

func (p *Parser) parseStatement() Statement {
	var (
		st  Statement
		err error
	)
	switch p.lex.CurToken().Token {
	case lexer.LET:
		st, err = p.parseLet()
	case lexer.RETURN:
		st, err = p.parseReturnStatement()
	// closure in Statements
	case lexer.BRLEFT:
		st = p.parseBlockStatement()
	default:
		st, err = p.parserExpressionStatement()
	}

	if err != nil {
		if !errors.As(err, &ParsingError{}) {
			panic(err)
		}

		p.Errors = append(p.Errors, err.(ParsingError))
	}

	for p.lex.PeekToken().Token == lexer.SCOLUMN {
		p.lex.Advance()
	}

	return st
}

func (p *Parser) isCurToken(token lexer.TokenType) bool {
	return p.lex.CurToken().Token == token
}

func (p *Parser) parserExpressionStatement() (Statement, error) {
	var err error
	st := ExpressionStatement{Tok: p.lex.CurToken()}

	st.Expr, err = p.parseExpression(LOWEST)
	return st, err
}

func (p *Parser) parseExpression(precedence int) (Expression, error) {
	prefix, ok := p.prefixParseFn[p.lex.CurToken().Token]
	if !ok {
		return nil, NewParsingError("missing Prefix parser", p.lex.CurToken())
	}

	left, err := prefix()
	if err != nil {
		return nil, err
	}

	for !p.isCurToken(lexer.SCOLUMN) && precedence < p.precedence(p.lex.PeekToken().Token) {
		p.lex.Advance()
		infix, ok := p.infixParseFn[p.lex.CurToken().Token]
		if !ok {
			return nil, NewParsingError("missing infix parser", p.lex.CurToken())
		}

		left, err = infix(left)
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func (p *Parser) precedence(token lexer.TokenType) int {
	switch token {
	case lexer.ASSIGN:
		return ASSIGN
	case lexer.DVERTLINE:
		return OR
	case lexer.DAMPERSAND:
		return AND
	case lexer.LT, lexer.LTE, lexer.GTE, lexer.GT, lexer.EQ, lexer.NEQ:
		return RELATIONAL
	case lexer.PLUS, lexer.HYPHEN:
		return ADDITION
	case lexer.SLASH, lexer.ASTERISK:
		return MULTIPLICATION
	case lexer.BANG:
		return PREFIX
	case lexer.BLEFT:
		return CALL
	}

	return LOWEST
}
