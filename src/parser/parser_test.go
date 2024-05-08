package parser

import (
	"bytes"
	"fmt"
	"github.com/charkpep/yad/src/lexer"
	"github.com/charkpep/yad/src/utils"
	"regexp"
	"testing"
)

func AssertRoot(t *testing.T, root Node) *RootNode {
	t.Helper()
	rootNode, ok := root.(*RootNode)
	if !ok {
		t.Errorf("expected root node, got %T\n", root)
	}

	return rootNode
}

func NewFuncExpression(idents []string, body []lexer.Token) []lexer.Token {
	fn := []lexer.Token{
		{
			Token:   lexer.FUNC,
			Literal: "fn",
		},
		{
			Token:   lexer.BLEFT,
			Literal: "(",
		},
	}

	for i, ident := range idents {
		fn = append(fn, lexer.Token{
			Token:   lexer.IDENT,
			Literal: ident,
		})
		if i+1 != len(idents) {
			fn = append(fn, lexer.Token{
				Token:   lexer.COMA,
				Literal: ",",
			})
		}
	}

	fn = append(fn, lexer.Token{
		Token:   lexer.BRIGHT,
		Literal: ")",
	}, lexer.Token{
		Token:   lexer.BRLEFT,
		Literal: "{",
	})

	fn = append(fn, body...)
	fn = append(fn, lexer.Token{
		Token:   lexer.BRRIGHT,
		Literal: "}",
	})

	return fn
}

func AssertExpressionStatement(t *testing.T, stmt Statement) ExpressionStatement {
	t.Helper()
	expressionStmt, ok := stmt.(ExpressionStatement)
	if !ok {
		t.Errorf("expected expresison statement node, got %T\n", stmt)
	}

	return expressionStmt
}

func AssertLetStatement(t *testing.T, stmt Statement) LetStatement {
	t.Helper()
	letStmt, ok := stmt.(LetStatement)
	if !ok {
		t.Errorf("expected root node, got %T\n", stmt)
	}

	return letStmt
}

func TestParseLetStatements(t *testing.T) {
	type tt struct {
		i string
		o *regexp.Regexp
	}

	ts := []tt{
		{
			i: `let x = 10;
				let y = 10;`,
			o: regexp.MustCompile(`(let [xy]=.*;\n){2}`),
		},
		{
			i: `let x = 10;
				let y = 10;`,
			o: regexp.MustCompile(`(let [xy]=10;\n){2}`),
		},
	}

	for i, test := range ts {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			p := NewParser(bytes.NewBuffer([]byte(test.i)))
			root, err := p.Parse()
			if err != nil {
				t.Error(err)
			}

			if !test.o.Match([]byte(root.String())) {
				t.Errorf("expected %s, got %q", test.o, root.String())
			}
		})
	}

}

func TestParseInfix(t *testing.T) {
	type tt struct {
		i lexer.ILexer
		o string
	}

	ts := []tt{
		{
			i: utils.NewMockLexer([]lexer.Token{{
				Token:   lexer.NUMBER,
				Line:    0,
				Column:  0,
				Literal: "5",
			}, {
				Token:   lexer.PLUS,
				Line:    0,
				Column:  0,
				Literal: "+",
			}, {
				Token:   lexer.NUMBER,
				Line:    0,
				Column:  0,
				Literal: "5",
			}}),
			o: "(5 + 5)\n",
		},
		{
			i: utils.NewMockLexer([]lexer.Token{{
				Token:   lexer.NUMBER,
				Literal: "5",
			}, {
				Token:   lexer.ASTERISK,
				Literal: "*",
			}, {
				Token:   lexer.NUMBER,
				Literal: "5",
			}, {

				Token:   lexer.HYPHEN,
				Literal: "-",
			}, {

				Token:   lexer.NUMBER,
				Literal: "5",
			}}),
			o: "((5 * 5) - 5)\n",
		},
		{
			i: utils.NewMockLexer([]lexer.Token{{
				Token:   lexer.NUMBER,
				Literal: "5",
			}, {
				Token:   lexer.ASTERISK,
				Literal: "*",
			}, {
				Token:   lexer.NUMBER,
				Literal: "5",
			}, {

				Token:   lexer.HYPHEN,
				Literal: "*",
			}, {

				Token:   lexer.NUMBER,
				Literal: "5",
			}, {
				Token:   lexer.HYPHEN,
				Literal: "-",
			}, {
				Token:   lexer.NUMBER,
				Literal: "5",
			}}),
			o: "(((5 * 5) * 5) - 5)\n",
		},
	}

	for _, test := range ts {
		p := NewParserFromLexer(test.i)
		root, err := p.Parse()
		if err != nil {
			t.Error(err)
		}

		if root.String() != test.o {
			t.Errorf("expected %q, got %q", test.o, root.String())
		}

	}

}

func TestParseExpression(t *testing.T) {
	type tt struct {
		i lexer.ILexer
		o string
	}

	tcs := []tt{
		{
			i: utils.NewMockLexer([]lexer.Token{
				{
					Token:   lexer.BANG,
					Literal: "!",
				},
				{
					Token:   lexer.NUMBER,
					Line:    0,
					Column:  0,
					Literal: "1",
				}}),
			o: "!(1)\n",
		},
		{
			i: utils.NewMockLexer([]lexer.Token{
				{
					Token:   lexer.BANG,
					Literal: "-",
				},
				{
					Token:   lexer.NUMBER,
					Literal: "1",
				}}),
			o: "-(1)\n",
		},
		{
			i: utils.NewMockLexer([]lexer.Token{
				{
					Token:   lexer.BRLEFT,
					Literal: "{",
				},
				{
					Token:   lexer.BRRIGHT,
					Literal: "}",
				},
			}),
			o: "{\n}\n",
		},
		{
			i: utils.NewMockLexer(NewFuncExpression([]string{}, []lexer.Token{})),
			o: "fn () {\n}\n",
		},
		{
			i: utils.NewMockLexer(NewFuncExpression([]string{"a", "b", "c"}, []lexer.Token{
				{
					Token:   lexer.LET,
					Literal: "let",
				},
				{
					Token:   lexer.IDENT,
					Literal: "a",
				},
				{
					Token:   lexer.ASSIGN,
					Literal: "=",
				},
				{
					Token:   lexer.NUMBER,
					Literal: "1",
				},
			})),
			o: "fn (a,b,c) {\nlet a=1;}\n",
		},
		{
			i: utils.NewMockLexer([]lexer.Token{
				{
					Token:   lexer.IDENT,
					Literal: "call",
				},
				{
					Token:   lexer.BLEFT,
					Literal: "(",
				},
				{
					Token:   lexer.BRIGHT,
					Literal: ")",
				},
			}),
			o: "call()\n",
		},
		{
			i: utils.NewMockLexer([]lexer.Token{
				{
					Token:   lexer.IDENT,
					Literal: "call",
				},
				{
					Token:   lexer.BLEFT,
					Literal: "(",
				},
				{
					Token:   lexer.IDENT,
					Literal: "a",
				},
				{
					Token:   lexer.COMA,
					Literal: ",",
				},
				{
					Token:   lexer.IDENT,
					Literal: "b",
				},
				{
					Token:   lexer.BRIGHT,
					Literal: ")",
				},
			}),
			o: "call(a,b)\n",
		},
		{
			i: utils.NewMockLexer([]lexer.Token{
				{
					Token:   lexer.TRUE,
					Literal: "true",
				},
				{
					Token:   lexer.FALSE,
					Literal: "true",
				},
			}),
			o: "true\nfalse\n",
		},
	}

	for i, test := range tcs {
		t.Log(test.i)
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			parser := NewParserFromLexer(test.i)
			node, err := parser.Parse()
			if err != nil {
				t.Error(err)
			}

			if len(parser.Errors) != 0 {
				t.Error(parser.Errors)
			}

			if node.String() != test.o {
				t.Errorf("expected %q, got %q", test.o, node.String())
			}
		})
	}
}

func TestParseBlockExpression(t *testing.T) {
	test := utils.NewMockLexer([]lexer.Token{
		{
			Token:   lexer.BRLEFT,
			Literal: "{",
		},
		{
			Token:   lexer.LET,
			Literal: "let",
		},
		{
			Token:   lexer.IDENT,
			Literal: "i",
		},
		{
			Token:   lexer.ASSIGN,
			Literal: "=",
		},
		{
			Token:   lexer.NUMBER,
			Literal: "1",
		},
		{
			Token:   lexer.BRRIGHT,
			Literal: "}",
		},
	})

	p := NewParserFromLexer(test)
	st := p.parseStatement()
	if len(p.Errors) != 0 {
		t.Errorf("unexpected error %v", p.Errors)
	}

	stmt, ok := st.(BlockStatement)
	if !ok {
		t.Errorf("expected block expression, got %T", stmt)
	}

	if len(stmt.statements) != 1 {
		t.Errorf("expected 1 statement, got %d", len(stmt.statements))
	}

	AssertLetStatement(t, stmt.statements[0])
}

func TestParseIfExpression(t *testing.T) {
	lex := utils.NewMockLexer([]lexer.Token{
		{
			Token:   lexer.IF,
			Literal: "if",
		},
		{
			Token:   lexer.BLEFT,
			Literal: "(",
		},
		{
			Token:   lexer.NUMBER,
			Literal: "1",
		},
		{
			Token:   lexer.EQ,
			Literal: "==",
		},
		{
			Token:   lexer.NUMBER,
			Literal: "1",
		},
		{
			Token:   lexer.BRIGHT,
			Literal: ")",
		},
		{
			Token:   lexer.BRLEFT,
			Literal: "{",
		},
		{
			Token:   lexer.LET,
			Literal: "let",
		},
		{
			Token:   lexer.IDENT,
			Literal: "a",
		},
		{
			Token:   lexer.ASSIGN,
			Literal: "=",
		},
		{
			Token:   lexer.NUMBER,
			Literal: "10",
		},
		{
			Token:   lexer.BRRIGHT,
			Literal: "}",
		},
	})
	condition := &InfixExpression{
		Left: IntegerExpression{
			Val: 1,
		},
		Operator: lexer.Token{
			Token:   lexer.EQ,
			Literal: "==",
		},
		Right: IntegerExpression{
			Val: 1,
		},
	}
	p := NewParserFromLexer(lex)
	root, err := p.Parse()
	if err != nil {
		t.Error(err)
	}

	if len(p.Errors) != 0 {
		t.Error(p.Errors)
	}

	rootNode := AssertRoot(t, root)
	if len(rootNode.Statements) != 1 {
		t.Errorf("expected 1 statement, got %d\n", len(rootNode.Statements))
	}

	expr := AssertExpressionStatement(t, rootNode.Statements[0])
	ifExpr, ok := expr.Expr.(IfExpression)
	if !ok {
		t.Errorf("expected if expression, got %T", ifExpr)
	}

	if !AssertNodes(t, ifExpr.condition, condition) {
		t.Errorf("expected %q, got %q", condition, ifExpr.condition)
	}

	if len(ifExpr.consequence.statements) != 1 {
		t.Errorf("expected %d, got %d Statements", 1, len(ifExpr.consequence.statements))
	}

	AssertLetStatement(t, ifExpr.consequence.statements[0])
	if ifExpr.alternative != nil {
		t.Errorf("expected alternative to be nil")
	}

}

// TODO: add error handling tests
