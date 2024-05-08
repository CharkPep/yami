package utils

import "github.com/charkpep/yad/src/lexer"

type LexerMock struct {
	tokens   []lexer.Token
	curToken int
}

func NewMockLexer(tokens []lexer.Token) lexer.ILexer {
	return &LexerMock{
		tokens: tokens,
	}
}

func (l *LexerMock) Advance() (lexer.Token, error) {
	l.curToken++
	return l.CurToken(), nil
}

func (l *LexerMock) CurToken() lexer.Token {
	if l.curToken >= len(l.tokens) {
		return lexer.NewToken(lexer.EOF, 0, 0, "")
	}

	return l.tokens[l.curToken]
}

func (l *LexerMock) PeekToken() lexer.Token {
	if l.curToken+1 >= len(l.tokens) {
		return lexer.NewToken(lexer.EOF, 0, 0, "")
	}

	return l.tokens[l.curToken+1]
}
