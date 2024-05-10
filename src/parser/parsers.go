package parser

import (
	"fmt"
	"github.com/charkpep/yad/src/lexer"
	"strconv"
)

func (p *Parser) ParseNumber() (Expression, error) {
	token := p.lex.CurToken()
	it := IntegerExpression{
		token: token,
	}

	var err error
	it.Val, err = strconv.ParseInt(token.Literal, 10, 64)
	if err != nil {
		return nil, err
	}

	return it, nil
}

func (p *Parser) parseIdentifier() (Expression, error) {
	literal := p.lex.CurToken()
	if literal.Token != lexer.IDENT {
		return nil, NewParsingError("expected Identifier", p.lex.CurToken())
	}
	return IdentifierExpression{
		Identifier: literal,
	}, nil
}

func (p *Parser) ParseInfix(expression Expression) (Expression, error) {
	infix := InfixExpression{
		Operator: p.lex.CurToken(),
		Left:     expression,
	}

	var err error
	precedence := p.precedence(p.lex.CurToken().Token)
	p.lex.Advance()
	infix.Right, err = p.parseExpression(precedence)
	return &infix, err
}

func (p *Parser) ParsePrefix() (Expression, error) {
	var err error
	prefix := PrefixExpression{
		Prefix: p.lex.CurToken(),
	}

	p.lex.Advance()
	prefix.Expr, err = p.parseExpression(PREFIX)
	return prefix, err
}

func (p *Parser) ParseGroupedExpression() (Expression, error) {

	p.lex.Advance()
	g, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	p.lex.Advance()
	if p.lex.CurToken().Token != lexer.BRIGHT {
		return nil, NewParsingError("expected closing bracket", p.lex.CurToken())
	}

	return g, err
}

func (p *Parser) parseLet() (Statement, error) {
	literal := p.lex.CurToken()
	statement := LetStatement{
		Literal: literal,
	}

	p.lex.Advance()
	identifier, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	statement.Identifier = identifier.(IdentifierExpression)
	p.lex.Advance()
	if p.lex.CurToken().Token != lexer.ASSIGN {
		return nil, NewParsingError("invalid token encountered", p.lex.CurToken())
	}

	p.lex.Advance()
	statement.Expression, err = p.parseExpression(LOWEST)
	return statement, err
}

func (p *Parser) parseBlockStatement() Statement {
	block := BlockStatement{
		token: p.lex.CurToken(),
	}

	p.lex.Advance()
	for !p.isCurToken(lexer.EOF) && !p.isCurToken(lexer.BRRIGHT) {
		st := p.parseStatement()
		if st != nil {
			block.Statements = append(block.Statements, st)
		}
		p.lex.Advance()
	}

	if !p.isCurToken(lexer.BRRIGHT) {
		p.Errors = append(p.Errors, NewParsingError("expected closing bracket, got EOF", block.token))
		return nil
	}

	return block
}

func (p *Parser) parseIfExpression() (Expression, error) {
	var err error
	ifExpr := IfExpression{
		token:       p.lex.CurToken(),
		Consequence: BlockStatement{},
	}

	p.lex.Advance()
	if p.lex.CurToken().Token == lexer.BLEFT {
		p.lex.Advance()
	}

	ifExpr.Condition, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if ifExpr.Condition == nil {
		return nil, NewParsingError("undefined Condition", ifExpr.token)
	}

	// So far
	//(1 == 1)
	//      ^
	p.lex.Advance()
	if p.lex.CurToken().Token == lexer.BRIGHT {
		p.lex.Advance()
	}

	consequence := p.parseBlockStatement()
	if consequence == nil {
		return nil, NewParsingError("undefined required Consequence of if expression", p.lex.CurToken())
	}

	ifExpr.Consequence = consequence.(BlockStatement)
	if p.lex.PeekToken().Token == lexer.ELSE {
		p.lex.Advance()
		p.lex.Advance()
		alternative := p.parseBlockStatement()
		if alternative == nil {
			return nil, NewParsingError("expected Alternative, got none", ifExpr.token)
		}

		alternativeBlock := alternative.(BlockStatement)
		ifExpr.Alternative = &alternativeBlock
	}

	return ifExpr, err
}

func (p *Parser) ParseFuncExpression() (Expression, error) {
	fn := FuncExpression{
		token: p.lex.CurToken(),
	}

	p.lex.Advance()
	if !p.isCurToken(lexer.BLEFT) {
		return nil, NewParsingError("expected (", p.lex.CurToken())
	}

	// First element in Args
	p.lex.Advance()
	if !p.isCurToken(lexer.BRIGHT) {
		args, err := p.parseComaSeparatedExpressions()
		if err != nil {
			return nil, err
		}

		for _, arg := range args {
			ident, ok := arg.(IdentifierExpression)
			if !ok {
				return nil, NewParsingError("expected Identifier", arg.Token())
			}
			fn.Args = append(fn.Args, ident)
		}

		p.lex.Advance()
	}

	if !p.isCurToken(lexer.BRIGHT) {
		return nil, NewParsingError("expected )", p.lex.CurToken())
	}

	p.lex.Advance()
	body := p.parseBlockStatement()
	if body == nil {
		return nil, NewParsingError("if Body is undefined", p.lex.CurToken())
	}

	fn.Body = body.(BlockStatement)
	return fn, nil
}

func (p *Parser) parseComaSeparatedExpressions() ([]Expression, error) {
	var expressions []Expression
	expr, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	expressions = append(expressions, expr)
	for p.lex.PeekToken().Token == lexer.COMA {
		p.lex.Advance()
		p.lex.Advance()
		expr, err = p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}

		expressions = append(expressions, expr)
	}

	return expressions, nil
}

func (p *Parser) parseCallExpression(fn Expression) (Expression, error) {
	call := CallExpression{
		token: p.lex.CurToken(),
		Call:  fn,
	}

	p.lex.Advance()
	if !p.isCurToken(lexer.BRIGHT) {
		var err error
		call.CallArgs, err = p.parseComaSeparatedExpressions()
		if err != nil {
			return nil, err
		}

		p.lex.Advance()
	}

	return call, nil
}

func (p *Parser) parseAssignExpression(ex Expression) (Expression, error) {
	if _, ok := ex.(IdentifierExpression); !ok {
		return nil, NewParsingError(fmt.Sprintf("expected Identifier, got %T\n", ex), p.lex.CurToken())
	}

	assign := AssignExpression{
		token:      p.lex.CurToken(),
		Identifier: ex.(IdentifierExpression),
	}

	p.lex.Advance()
	var err error
	assign.Val, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	return assign, err
}

func (p *Parser) parseBoolExpression() (Expression, error) {
	var val bool
	if p.lex.CurToken().Token == lexer.TRUE {
		val = true
	}

	return BoolExpression{
		token: p.lex.CurToken(),
		Val:   val,
	}, nil
}

func (p *Parser) parseReturnStatement() (Statement, error) {
	rt := ReturnStatement{
		token: p.lex.CurToken(),
	}

	p.lex.Advance()
	var err error
	rt.ReturnExpr, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	return rt, err
}
