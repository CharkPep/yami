package parser

import (
	"fmt"
	"github.com/charkpep/yad/src/lexer"
	"strconv"
)

func (p *Parser) ParseNumber() (Expression, error) {
	token := p.curToken
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
	literal := p.curToken
	if literal.Token != lexer.IDENT {
		return nil, NewParsingError("expected Identifier", p.curToken)
	}
	return IdentifierExpression{
		Identifier: literal,
	}, nil
}

func (p *Parser) ParseInfix(expression Expression) (Expression, error) {
	infix := InfixExpression{
		Operator: p.curToken,
		Left:     expression,
	}

	var err error
	precedence := p.precedence(p.curToken.Token)
	p.read()
	infix.Right, err = p.parseExpression(precedence)
	return &infix, err
}

func (p *Parser) ParsePrefix() (Expression, error) {
	var err error
	prefix := PrefixExpression{
		Prefix: p.curToken,
	}

	p.read()
	prefix.Expr, err = p.parseExpression(PREFIX)
	return prefix, err
}

func (p *Parser) ParseGroupedExpression() (Expression, error) {

	p.read()
	g, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	p.read()
	if p.curToken.Token != lexer.BRIGHT {
		return nil, NewParsingError("expected closing bracket", p.curToken)
	}

	return g, err
}

func (p *Parser) parseLet() (Statement, error) {
	literal := p.curToken
	statement := LetStatement{
		Literal: literal,
	}

	p.read()
	identifier, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	statement.Identifier = identifier.(IdentifierExpression)
	p.read()
	if p.curToken.Token != lexer.ASSIGN {
		return nil, NewParsingError("invalid token encountered", p.curToken)
	}

	p.read()
	statement.Expression, err = p.parseExpression(LOWEST)
	return statement, err
}

func (p *Parser) parseBlockStatement() (Statement, error) {
	block := BlockStatement{
		token:      p.curToken,
		Statements: make([]Statement, 0),
	}

	p.read()
	for !p.isCurToken(lexer.EOF) && !p.isCurToken(lexer.BRRIGHT) {
		st, err := p.parseStatement()
		if err != nil {
			return nil, err
		}

		if st != nil {
			block.Statements = append(block.Statements, st)
		}
		p.read()
	}

	if !p.isCurToken(lexer.BRRIGHT) {
		p.Errors = append(p.Errors, NewParsingError("expected closing bracket, got EOF", block.token))
		return nil, nil
	}

	return block, nil
}

func (p *Parser) parseIfExpression() (Expression, error) {
	var err error
	ifExpr := IfExpression{
		token:       p.curToken,
		Consequence: BlockStatement{},
	}

	p.read()
	if p.curToken.Token == lexer.BLEFT {
		p.read()
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
	p.read()
	if p.curToken.Token == lexer.BRIGHT {
		p.read()
	}

	consequence, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}

	if consequence == nil {
		return nil, NewParsingError("undefined required Consequence of if expression", p.curToken)
	}

	ifExpr.Consequence = consequence.(BlockStatement)
	if p.peekToken.Token == lexer.ELSE {
		p.read()
		p.read()
		alternative, err := p.parseBlockStatement()
		if err != nil {
			return nil, err
		}

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
		token: p.curToken,
	}

	p.read()
	if !p.isCurToken(lexer.BLEFT) {
		return nil, NewParsingError("expected (", p.curToken)
	}

	// First element in Args
	p.read()
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

		p.read()
	}

	if !p.isCurToken(lexer.BRIGHT) {
		return nil, NewParsingError("expected )", p.curToken)
	}

	p.read()
	body, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	if body == nil {
		return nil, NewParsingError("if Body is undefined", p.curToken)
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
	for p.peekToken.Token == lexer.COMA {
		p.read()
		p.read()
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
		token: p.curToken,
		Call:  fn,
	}

	p.read()
	if !p.isCurToken(lexer.BRIGHT) {
		var err error
		call.CallArgs, err = p.parseComaSeparatedExpressions()
		if err != nil {
			return nil, err
		}

		p.read()
	}

	return call, nil
}

func (p *Parser) parseAssignExpression(ex Expression) (Expression, error) {
	switch ex.(type) {
	case IdentifierExpression:
	case IndexExpression:
	default:
		return nil, NewParsingError(fmt.Sprintf("expected Identifier, got %T\n", ex), p.curToken)
	}

	assign := AssignExpression{
		token:      p.curToken,
		Identifier: ex,
	}

	p.read()
	var err error
	assign.Val, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	return assign, err
}

func (p *Parser) parseBoolExpression() (Expression, error) {
	var val bool
	if p.curToken.Token == lexer.TRUE {
		val = true
	}

	return BoolExpression{
		token: p.curToken,
		Val:   val,
	}, nil
}

func (p *Parser) parseReturnStatement() (Statement, error) {
	rt := ReturnStatement{
		token:      p.curToken,
		ReturnExpr: NilExpression{token: p.curToken},
	}

	if p.peekToken.Token != lexer.BRRIGHT {
		var err error
		p.read()
		rt.ReturnExpr, err = p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
	}

	return rt, nil
}

func (p *Parser) parseStringExpression() (Expression, error) {
	return StringExpression{
		tok: p.curToken,
		Val: p.curToken.Literal,
	}, nil
}

func (p *Parser) parseIndexExpression(expr Expression) (Expression, error) {
	idx := IndexExpression{
		token: p.curToken,
		Of:    expr,
	}
	p.read()
	var err error
	idx.Idx, err = p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	p.read()

	return idx, err
}

func (p *Parser) parseArrayExpression() (Expression, error) {
	arr := ArrayExpression{
		token: p.curToken,
	}

	p.read()
	if p.curToken.Token != lexer.SBRIGHT {
		var err error
		arr.Arr, err = p.parseComaSeparatedExpressions()
		if err != nil {
			return nil, err
		}
		p.read()
	}

	return arr, nil
}

func (p *Parser) parseHashMap() (Expression, error) {
	mp := HashMapExpression{
		token: p.curToken,
	}

	p.read()
	if p.curToken.Token != lexer.BRRIGHT {
		var err error
		mp.Map, err = p.parseComaSeparatedPairs()
		if err != nil {
			return nil, err
		}
		p.read()
	}

	return mp, nil
}

func (p *Parser) parseComaSeparatedPairs() (map[Expression]Expression, error) {
	expressions := make(map[Expression]Expression)
	key, val, err := p.parserPair()
	if err != nil {
		return nil, err
	}

	expressions[key] = val
	for p.peekToken.Token == lexer.COMA {
		p.read()
		p.read()
		key, val, err = p.parserPair()
		if err != nil {
			return nil, err
		}

		expressions[key] = val
	}

	return expressions, nil
}

func (p *Parser) parserPair() (Expression, Expression, error) {
	key, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, nil, err
	}

	if p.peekToken.Token != lexer.COLON {
		return nil, nil, NewParsingError("expected proceeding column", p.curToken)
	}

	p.read()
	p.read()

	val, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, nil, err
	}

	return key, val, err
}
