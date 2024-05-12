package lexer

import (
	"bytes"
	"fmt"
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
				Token:   SCOLON,
				Literal: ";",
			},
		},
	}

	for _, test := range ts {
		lexer := New(bytes.NewBufferString(test.i))
		tok := Token{}
		err := lexer.Read(&tok)
		if err != nil {
			t.Error(err)
		}

		if tok.Token != test.out.Token || tok.Literal != test.out.Literal {
			t.Errorf("expected %v, got %v", test.out, tok)
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
			i: `arr= []; 5/2 true, false "hello world!"`,
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
					Token:   SCOLON,
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
				{
					Token:   STRING,
					Literal: "hello world!",
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
			i: `5*100!=501 || 1 == 1 && 1 == 1 < > <= >= if() {} else [] [1,1,1] & | << >> { "a" : "b" }`,
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
					Token:   OR,
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
					Token:   AND,
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
				{
					Token:   SBLEFT,
					Literal: "[",
				},
				{
					Token:   SBRIGHT,
					Literal: "]",
				},
				{
					Token:   SBLEFT,
					Literal: "[",
				},
				{
					Token:   NUMBER,
					Literal: "1",
				},
				{
					Token:   COMA,
					Literal: ",",
				},
				{
					Token:   NUMBER,
					Literal: "1",
				},
				{
					Token:   COMA,
					Literal: ",",
				},
				{
					Token:   NUMBER,
					Literal: "1",
				},
				{
					Token:   SBRIGHT,
					Literal: "]",
				},
				{
					Token:   BAND,
					Literal: "&",
				},
				{
					Token:   BOR,
					Literal: "|",
				},
				{
					Token:   BLSHIFT,
					Literal: "<<",
				},
				{
					Token:   BRSHIFT,
					Literal: ">>",
				},
				{
					Token:   BRLEFT,
					Literal: "{",
				},
				{
					Token:   STRING,
					Literal: "a",
				},
				{
					Token:   COLON,
					Literal: ":",
				},
				{
					Token:   STRING,
					Literal: "b",
				},
				{
					Token:   BRRIGHT,
					Literal: "}",
				},
			},
		}}

	for _, test := range ts {
		lexer := New(bytes.NewBufferString(test.i))
		for i, expected := range test.out {
			tok := Token{}
			err := lexer.Read(&tok)
			if err != nil {
				t.Error(err)
			}

			if expected.Token != tok.Token || expected.Literal != tok.Literal {
				t.Errorf("expected %q, got %q, token %d", expected, tok, i)
			}

		}
	}
}

func TestLexerLineAndColumnCount(t *testing.T) {
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
					Token:   SCOLON,
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
					Token:   SCOLON,
					Line:    1,
					Column:  22,
					Literal: ";",
				},
				{
					Token:   EOF,
					Line:    1,
					Column:  22,
					Literal: "",
				},
			},
		},
	}

	for i, test := range ts {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			lexer := New(bytes.NewBufferString(test.i))
			for i, expected := range test.o {
				tok := Token{}
				err := lexer.Read(&tok)
				if err != nil {
					t.Error(err)
				}

				if tok != expected {
					t.Errorf("token %d: expected %v, got %v", i, expected, tok)
				}

			}

		})

	}
}
