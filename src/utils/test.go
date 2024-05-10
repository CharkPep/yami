package utils

import "github.com/charkpep/yad/src/lexer"

type LexerMock struct {
	tokens    []lexer.Token
	peekToken int
}

func NewMockLexer(tokens []lexer.Token) lexer.TokenReader {
	return &LexerMock{
		tokens: tokens,
	}
}

func (l *LexerMock) Read(t *lexer.Token) error {
	if l.peekToken+1 > len(l.tokens) {
		t.Token = lexer.EOF
		t.Literal = ""
		return nil
	}

	if t == nil {
		t = &l.tokens[l.peekToken]
	}

	*t = l.tokens[l.peekToken]
	l.peekToken++
	return nil
}
