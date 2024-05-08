package lexer

import (
	"bytes"
	"testing"
)

func TestSingeToken(t *testing.T) {
	type tt struct {
		i   string
		out Token
	}

	ts := []tt{
		{
			i: "=",
			out: Token{
				Token:   ASSIGN,
				Literal: "=",
			},
		},
		{
			i: ";",
			out: Token{
				Token:   SCOLUMN,
				Literal: ";",
			},
		},
		{
			i: `"`,
			out: Token{
				Token:   DQUOTE,
				Literal: `"`,
			},
		},
	}

	for _, test := range ts {
		lexer := New(bytes.NewBufferString(test.i))
		token, err := lexer.Advance()
		if err != nil {
			t.Error(err)
		}

		if token.Token != test.out.Token || token.Literal != test.out.Literal {
			t.Errorf("expected %v, got %v", test.out, token)
		}
	}

}

func TestMultipleTokens(t *testing.T) {
	type tt struct {
		i   string
		out []Token
	}

	ts := []tt{
		{
			i: "arr= [];\n 5/2 true, false",
			out: []Token{
				{
					Token:   IDENT,
					Literal: "arr",
				},
				{
					Token:   ASSIGN,
					Literal: "=",
				},
				{
					Token:   SBLEFT,
					Literal: "[",
				},
				{
					Token:   SBRIGHT,
					Literal: "]",
				},
				{
					Token:   SCOLUMN,
					Literal: ";",
				},
				{
					Token:   NUMBER,
					Literal: "5",
				},
				{
					Token:   SLASH,
					Literal: "/",
				},
				{
					Token:   NUMBER,
					Literal: "2",
				},
				{
					Token:   TRUE,
					Literal: "true",
				},
				{
					Token:   COMA,
					Literal: ",",
				},
				{
					Token:   FALSE,
					Literal: "false",
				},
			},
		},
		{
			i: `//comment
//comment
//comment
//comment
let
`,
			out: []Token{
				{
					Token:   LET,
					Literal: "let",
				},
			},
		},
		{
			i: "5*100!=501 || 1 == 1 && 1 == 1 < > <= >= if() {} else",
			out: []Token{
				{
					Token:   NUMBER,
					Literal: "5",
				},
				{
					Token:   ASTERISK,
					Literal: "*",
				},
				{
					Token:   NUMBER,
					Literal: "100",
				},
				{
					Token:   NEQ,
					Literal: "!=",
				},
				{
					Token:   NUMBER,
					Literal: "501",
				},
				{
					Token:   DVERTLINE,
					Literal: "||",
				},
				{
					Token:   NUMBER,
					Literal: "1",
				},
				{
					Token:   EQ,
					Literal: "==",
				},
				{
					Token:   NUMBER,
					Literal: "1",
				},
				{
					Token:   DAMPERSAND,
					Literal: "&&",
				},

				{
					Token:   NUMBER,
					Literal: "1",
				},
				{
					Token:   EQ,
					Literal: "==",
				},
				{
					Token:   NUMBER,
					Literal: "1",
				},
				{
					Token:   LT,
					Literal: "<",
				},
				{
					Token:   GT,
					Literal: ">",
				},
				{
					Token:   LTE,
					Literal: "<=",
				},
				{
					Token:   GTE,
					Literal: ">=",
				},
				{
					Token:   IF,
					Literal: "if",
				},
				{
					Token:   BLEFT,
					Literal: "(",
				},
				{
					Token:   BRIGHT,
					Literal: ")",
				},
				{
					Token:   BRLEFT,
					Literal: "{",
				},
				{
					Token:   BRRIGHT,
					Literal: "}",
				},
				{
					Token:   ELSE,
					Literal: "else",
				},
			},
		}}

	for _, test := range ts {
		lexer := New(bytes.NewBufferString(test.i))
		for i, expected := range test.out {
			token, err := lexer.Advance()
			if err != nil {
				t.Error(err)
			}

			if expected.Token != token.Token {
				t.Errorf("expected %s, got %s, token %d", expected.Token, token.Token, i)
			}

			if expected.Token != lexer.CurToken().Token {
				t.Errorf("expected %s, got %s, token %d", expected.Token, lexer.CurToken().Token, i)
			}

		}
	}
}

func TestBufferedLexer_PeekToken(t *testing.T) {
	type tt struct {
		i    string
		cur  Token
		peek Token
	}

	tc := []tt{
		{
			i: "let a = 10",
			cur: Token{
				Token:   LET,
				Literal: "let",
			},
			peek: Token{
				Token:   IDENT,
				Line:    0,
				Column:  0,
				Literal: "a",
			},
		},
		{
			i: `
//comment
//comment
//comment
//comment
a b
`,
			cur: Token{
				Token:   IDENT,
				Literal: "a",
			},
			peek: Token{
				Token:   IDENT,
				Line:    0,
				Column:  0,
				Literal: "b",
			},
		},
	}

	for _, test := range tc {
		lexer := New(bytes.NewBufferString(test.i))
		if _, err := lexer.Advance(); err != nil {
			t.Error(err)
		}

		if token := lexer.CurToken(); token.Token != test.cur.Token || token.Literal != test.cur.Literal {
			t.Errorf("expected %+v, got %+v", test.cur, token)
		}

		if token := lexer.PeekToken(); token.Token != test.peek.Token || token.Literal != test.peek.Literal {
			t.Errorf("expected %+v, got %+v", test.peek, token)
		}

	}

}

func TestBufferedLexer_LineAndColumnCount(t *testing.T) {
	type tt struct {
		i string
		o []Token
	}

	ts := []tt{
		{
			i: "let a = 10;",
			o: []Token{
				{
					Token:   LET,
					Line:    0,
					Column:  3,
					Literal: "let",
				},
				{
					Token:   IDENT,
					Line:    0,
					Column:  5,
					Literal: "a",
				},
				{
					Token:   ASSIGN,
					Line:    0,
					Column:  7,
					Literal: "=",
				},
				{
					Token:   NUMBER,
					Line:    0,
					Column:  10,
					Literal: "10",
				},
				{
					Token:   SCOLUMN,
					Line:    0,
					Column:  11,
					Literal: ";",
				},
			},
		},
		{
			i: `//comment a
let abc_aaa ==  != 10;`,
			o: []Token{
				{
					Token:   LET,
					Line:    1,
					Column:  3,
					Literal: "let",
				},
				{
					Token:   IDENT,
					Line:    1,
					Column:  11,
					Literal: "abc_aaa",
				},
				{
					Token:   EQ,
					Line:    1,
					Column:  14,
					Literal: "==",
				},
				{
					Token:   NEQ,
					Line:    1,
					Column:  18,
					Literal: "!=",
				},
				{
					Token:   NUMBER,
					Line:    1,
					Column:  21,
					Literal: "10",
				},
				{
					Token:   SCOLUMN,
					Line:    1,
					Column:  22,
					Literal: ";",
				},
			},
		},
	}

	for _, test := range ts {
		lexer := New(bytes.NewBufferString(test.i))
		for i, expected := range test.o {
			tok, err := lexer.Advance()
			if err != nil {
				t.Error(err)
			}

			if tok != expected {
				t.Errorf("token %d: expected %v, got %v", i, expected, tok)
			}

		}

	}

}
