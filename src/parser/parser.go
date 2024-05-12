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
	BIN_OR
	XOR
	BIN_AND
	BIN_SHIFT
	ADDITION // +
	MULTIPLICATION
	PREFIX // -5 !5
	IDX    // a[1]
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
	lex lexer.TokenReader

	prefixParseFn map[lexer.TokenType]prefixParseFn
	infixParseFn  map[lexer.TokenType]infixParseFn

	curToken  lexer.Token
	peekToken lexer.Token
	Errors    []ParsingError
}

func (p *Parser) registerPrefixFunc(token lexer.TokenType, fn prefixParseFn) {
	if _, ok := p.prefixParseFn[token]; ok {
		panic(fmt.Sprintf("parser for %q, already exists", token))
	}
	p.prefixParseFn[token] = fn
}

func (p *Parser) registerInfixFunc(token lexer.TokenType, fn infixParseFn) {
	if _, ok := p.infixParseFn[token]; ok {
		panic(fmt.Sprintf("parser for %q, already exists", token))
	}
	p.infixParseFn[token] = fn
}

func (p *Parser) registerInfixesFunc(fn infixParseFn, tokens ...lexer.TokenType) {
	for _, t := range tokens {
		p.infixParseFn[t] = fn
	}
}

func NewParserFromLexer(lex lexer.TokenReader) *Parser {
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
	p.registerPrefixFunc(lexer.SBLEFT, p.parseArrayExpression)
	p.registerPrefixFunc(lexer.TRUE, p.parseBoolExpression)
	p.registerPrefixFunc(lexer.FALSE, p.parseBoolExpression)
	p.registerPrefixFunc(lexer.STRING, p.parseStringExpression)
	p.registerPrefixFunc(lexer.BRLEFT, p.parseHashMap)
	p.registerInfixFunc(lexer.SBLEFT, p.parseIndexExpression)
	p.registerInfixesFunc(p.ParseInfix, lexer.PLUS, lexer.HYPHEN, lexer.SLASH, lexer.ASTERISK, lexer.EQ, lexer.NEQ,
		lexer.OR, lexer.AND, lexer.GT, lexer.GTE, lexer.LT, lexer.LTE, lexer.BOR, lexer.BAND, lexer.BLSHIFT, lexer.BRSHIFT)

	return p
}

func (p *Parser) read() error {
	p.curToken = p.peekToken
	if err := p.lex.Read(&p.peekToken); err != nil {
		return err
	}

	return nil
}

func (p *Parser) Parse() (Node, error) {
	root := RootNode{}
	p.read()
	var err error
	for err = p.read(); err == nil && p.curToken.Token != lexer.ILLEGAL && p.curToken.Token != lexer.EOF; err = p.read() {
		st, err := p.parseStatement()
		if err != nil {
			return nil, err
		}

		if st == nil {
			continue
		}

		root.Statements = append(root.Statements, st)
	}

	if err != nil {
		return nil, err
	}

	return &root, nil
}

func (p *Parser) parseStatement() (Statement, error) {
	var (
		st  Statement
		err error
	)
	switch p.curToken.Token {
	case lexer.LET:
		st, err = p.parseLet()
	case lexer.RETURN:
		st, err = p.parseReturnStatement()
	case lexer.BRLEFT:
		st, err = p.parseBlockStatement()
	case lexer.SCOLON:
		break
	default:
		st, err = p.parserExpressionStatement()
	}

	if err != nil {
		if !errors.As(err, &ParsingError{}) {
			return nil, err
		}

		p.Errors = append(p.Errors, err.(ParsingError))
	}

	if p.peekToken.Token == lexer.SCOLON {
		if err = p.read(); err != nil {
			return nil, err
		}
	}

	return st, nil
}

func (p *Parser) isCurToken(token lexer.TokenType) bool {
	return p.curToken.Token == token
}

func (p *Parser) parserExpressionStatement() (Statement, error) {
	var err error
	st := ExpressionStatement{Tok: p.curToken}

	st.Expr, err = p.parseExpression(LOWEST)
	return st, err
}

func (p *Parser) parseExpression(precedence int) (Expression, error) {
	prefix, ok := p.prefixParseFn[p.curToken.Token]
	if !ok {
		return nil, NewParsingError("missing Prefix parser", p.curToken)
	}

	left, err := prefix()
	if err != nil {
		return nil, err
	}

	for !p.isCurToken(lexer.SCOLON) && precedence < p.precedence(p.peekToken.Token) {
		p.read()
		infix, ok := p.infixParseFn[p.curToken.Token]
		if !ok {
			return nil, NewParsingError("missing infix parser", p.curToken)
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
	case lexer.OR:
		return OR
	case lexer.AND:
		return AND
	case lexer.LT, lexer.LTE, lexer.GTE, lexer.GT, lexer.EQ, lexer.NEQ:
		return RELATIONAL
	case lexer.BAND:
		return BIN_AND
	case lexer.BOR:
		return BIN_OR
	case lexer.BLSHIFT, lexer.BRSHIFT:
		return BIN_SHIFT
	case lexer.PLUS, lexer.HYPHEN:
		return ADDITION
	case lexer.SLASH, lexer.ASTERISK:
		return MULTIPLICATION
	case lexer.BANG:
		return PREFIX
	case lexer.SBLEFT:
		return IDX
	case lexer.BLEFT:
		return CALL
	}

	return LOWEST
}
