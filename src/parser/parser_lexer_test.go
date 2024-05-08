package parser

import (
	"bytes"
	"fmt"
	"github.com/charkpep/yad/src/lexer"
	"reflect"
	"testing"
)

func AssertNodes(t *testing.T, a, b Node) bool {
	t.Helper()
	t.Logf("Asserting %T and %T, %+v, %+v\n", a, b, a, b)
	if reflect.ValueOf(a).Kind() != reflect.ValueOf(b).Kind() {
		t.Logf("Got different value kinds: (%v, %s), (%v, %s)\n", a, reflect.ValueOf(a).Kind(), b, reflect.ValueOf(b).Kind())
		return false
	}

	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		t.Logf("Got different value types: (%v, %T), (%v, %T)\n", a, a, b, b)
		return false
	}

	switch v := a.(type) {
	case *RootNode:
		rootB := AssertRoot(t, b)
		if len(v.Statements) != len(rootB.Statements) {
			t.Errorf("failed to assert root nodes, wrong number of Statements, got: %d and %d\n", len(v.Statements), len(b.(*RootNode).Statements))
			return false
		}

		for i, _ := range v.Statements {
			if !AssertNodes(t, v.Statements[i], rootB.Statements[i]) {
				return false
			}
		}
	case ExpressionStatement:
		if !AssertNodes(t, v.Expr, b.(ExpressionStatement).Expr) {
			return false
		}
	case IntegerExpression:
		if v.Val != b.(IntegerExpression).Val {
			t.Errorf("failed to assert integer nodes, %v, %v\n", v, b.(IntegerExpression))
			return false
		}
	case IdentifierExpression:
		if v.identifier.Literal != b.(IdentifierExpression).identifier.Literal {
			return false
		}
	case LetStatement:
		if !AssertNodes(t, v.Expression, b.(LetStatement).Expression) {
			return false
		}
	case *BlockStatement:
		if len(v.statements) != len(b.(*BlockStatement).statements) {
			t.Errorf("Length of Blockstatement is not equal: %d, %d\n", len(v.statements), len(b.(*BlockStatement).statements))
			return false
		}

		for i, _ := range v.statements {
			if !AssertNodes(t, v.statements[i], b.(*BlockStatement).statements[i]) {
				return false
			}
		}
	case BlockStatement:
		if len(v.statements) != len(b.(BlockStatement).statements) {
			t.Errorf("Length of Blockstatement is not equal: %d, %d\n", len(v.statements), len(b.(BlockStatement).statements))
			return false
		}

		for i, _ := range v.statements {
			if !AssertNodes(t, v.statements[i], b.(BlockStatement).statements[i]) {
				return false
			}
		}
	case IfExpression:
		if !AssertNodes(t, v.condition, b.(IfExpression).condition) || !AssertNodes(t, v.consequence, b.(IfExpression).consequence) {
			return false
		}

		if v.alternative != nil && !AssertNodes(t, v.alternative, b.(IfExpression).alternative) {
			return false
		}
	case *InfixExpression:
		if v.Operator.Token != b.(*InfixExpression).Operator.Token && v.Operator.Literal != b.(*InfixExpression).Operator.Literal {
			t.Errorf("failed to assert infixes, %v, %v\n", v, b.(*InfixExpression))
			return false
		}

		if !AssertNodes(t, v.Right, b.(*InfixExpression).Right) || !AssertNodes(t, v.Left, b.(*InfixExpression).Left) {
			return false
		}
	case PrefixExpression:
		if v.Prefix.Literal != b.(PrefixExpression).Prefix.Literal || v.Prefix.Token != b.(PrefixExpression).Prefix.Token {
			t.Errorf("failed to assert prefixes, got %v and %v\n", v.Prefix, b.(PrefixExpression).Prefix)
		}

		if !AssertNodes(t, v.Expr, b.(PrefixExpression).Expr) {
			return false
		}
	case FuncExpression:
		if len(v.args) != len(b.(FuncExpression).args) {
			t.Errorf("number of argumnets does not match\n")
			return false
		}
		for i, _ := range v.args {
			if !AssertNodes(t, v.args[i], b.(FuncExpression).args[i]) {
				return false
			}
		}

		if len(v.body.statements) != len(b.(FuncExpression).body.statements) {
			t.Errorf("number of Statements in body does not match\n")
			return false
		}

		for i, _ := range v.body.statements {
			return AssertNodes(t, v.body.statements[i], b.(FuncExpression).body.statements[i])
		}
	case AssignExpression:
		if !AssertNodes(t, v.Identifier, b.(AssignExpression).Identifier) {
			return false
		}

		if !AssertNodes(t, v.Val, b.(AssignExpression).Val) {
			return false
		}
	case ReturnStatement:
		if !AssertNodes(t, v.returnExpr, b.(ReturnStatement).returnExpr) {
			return false
		}

		if v.token.Token != b.(ReturnStatement).token.Token && v.token.Literal != b.(ReturnStatement).token.Literal {
			t.Logf("return statement tokens are different")
			return false
		}
	default:
		t.Errorf("Not supported type %T\n", v)
		return false
	}

	return true
}

func TestArithmeticAndLogicOperations(t *testing.T) {
	type tt struct {
		i string
		o interface{}
	}

	ts := []tt{
		{
			i: "1 + 1",
			o: &InfixExpression{
				Left: IntegerExpression{
					Val: 1,
				},
				Operator: lexer.Token{
					Token:   lexer.PLUS,
					Literal: "+",
				},
				Right: IntegerExpression{
					Val: 1,
				},
			},
		},
		{
			i: "1 + 1 + 2",
			o: &InfixExpression{
				Left: &InfixExpression{
					Left: IntegerExpression{
						Val: 1,
					},
					Operator: lexer.Token{
						Token:   lexer.PLUS,
						Literal: "+",
					},
					Right: IntegerExpression{
						Val: 1,
					},
				},
				Operator: lexer.Token{
					Token:   lexer.PLUS,
					Literal: "+",
				},
				Right: IntegerExpression{
					Val: 2,
				},
			},
		},
		{
			i: "1 / (10 - 1 * 10)",
			o: &InfixExpression{
				Left: IntegerExpression{
					Val: 1,
				},
				Operator: lexer.Token{
					Token:   lexer.SLASH,
					Literal: "/",
				},
				Right: &InfixExpression{
					Left: IntegerExpression{
						Val: 10,
					},
					Operator: lexer.Token{
						Token:   lexer.HYPHEN,
						Literal: "-",
					},
					Right: &InfixExpression{
						Left: IntegerExpression{
							Val: 1,
						},
						Operator: lexer.Token{
							Token:   lexer.ASTERISK,
							Literal: "*",
						},
						Right: IntegerExpression{
							Val: 10,
						},
					},
				},
			},
		},
		{
			i: "1 / (10 - 1 * 10 + 10)",
			o: &InfixExpression{
				Left: IntegerExpression{
					Val: 1,
				},
				Operator: lexer.Token{
					Token:   lexer.SLASH,
					Literal: "/",
				},
				Right: &InfixExpression{
					Left: &InfixExpression{
						Left: IntegerExpression{
							Val: 10,
						},
						Operator: lexer.Token{
							Token:   lexer.HYPHEN,
							Literal: "-",
						},
						Right: &InfixExpression{
							Left: IntegerExpression{
								Val: 1,
							},
							Operator: lexer.Token{
								Token:   lexer.ASTERISK,
								Literal: "*",
							},
							Right: IntegerExpression{
								Val: 10,
							},
						},
					},
					Operator: lexer.Token{
						Token:   lexer.HYPHEN,
						Literal: "+",
					},
					Right: IntegerExpression{
						Val: 10,
					},
				},
			},
		},
		{
			i: "1 == 1",
			o: &InfixExpression{
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
			},
		},
		{
			i: "(1 == 1 || 2 || 2) && 1 + 1 != 1",
			o: &InfixExpression{
				Left: &InfixExpression{
					Left: &InfixExpression{
						Left: &InfixExpression{
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
						},
						Operator: lexer.Token{
							Token:   lexer.DVERTLINE,
							Literal: "||",
						},
						Right: IntegerExpression{
							Val: 2,
						},
					},
					Operator: lexer.Token{
						Token:   lexer.DVERTLINE,
						Literal: "||",
					},
					Right: IntegerExpression{
						Val: 2,
					},
				},
				Operator: lexer.Token{
					Token:   lexer.DAMPERSAND,
					Literal: "&&",
				},
				Right: &InfixExpression{
					Left: &InfixExpression{
						Left: IntegerExpression{Val: 1},
						Operator: lexer.Token{
							Token:   lexer.PLUS,
							Literal: "+",
						},
						Right: IntegerExpression{Val: 1},
					},
					Operator: lexer.Token{
						Token:   lexer.NEQ,
						Literal: "!=",
					},
					Right: IntegerExpression{Val: 1},
				},
			},
		},
		{
			i: "!1 == -1",
			o: &InfixExpression{
				Left: PrefixExpression{
					Prefix: lexer.Token{
						Token:   lexer.BANG,
						Literal: "!",
					},
					Expr: IntegerExpression{
						Val: 1,
					},
				},
				Operator: lexer.Token{
					Token:   lexer.EQ,
					Literal: "==",
				},
				Right: PrefixExpression{
					Prefix: lexer.Token{
						Token:   lexer.HYPHEN,
						Literal: "-",
					},
					Expr: IntegerExpression{
						Val: 1,
					},
				},
			},
		},
	}

	for _, test := range ts {
		p := NewParser(bytes.NewBufferString(test.i))
		root, err := p.Parse()
		if err != nil {
			t.Error(err)
		}

		if len(p.Errors) != 0 {
			t.Error(p.Errors)
		}

		rootNode := AssertRoot(t, root)
		if len(rootNode.Statements) != 1 {
			t.Errorf("expected 1, got %d", len(rootNode.Statements))
		}

		stmt := AssertExpressionStatement(t, rootNode.Statements[0])
		if !AssertNodes(t, stmt.Expr, test.o.(Node)) {
			t.Errorf("Assertion failed")
		}
	}
}

func TestAstTreeWithConcreteLexer(t *testing.T) {
	type tt struct {
		i string
		o interface{}
	}

	ts := []tt{
		{
			i: `let a = 10
				let a = 10;
				`,
			o: &RootNode{
				Statements: []Statement{
					LetStatement{
						Literal: lexer.Token{
							Token:   lexer.LET,
							Literal: "let",
						},
						Identifier: IdentifierExpression{
							identifier: lexer.Token{
								Token:   lexer.IDENT,
								Literal: "a",
							},
						},
						Expression: IntegerExpression{
							Val: 10,
						},
					},
					LetStatement{
						Literal: lexer.Token{
							Token:   lexer.LET,
							Literal: "let",
						},
						Identifier: IdentifierExpression{
							identifier: lexer.Token{
								Token:   lexer.IDENT,
								Literal: "a",
							},
						},
						Expression: IntegerExpression{
							Val: 10,
						},
					},
				},
			},
		},
		{
			i: `
				let a = 10;
				let b = 5;
				if (a > b) {
					if b == 5 {
						let b = 10;
					}
					let a = 10;
				} else {
					let b = 5
				}
				`,
			o: &RootNode{
				Statements: []Statement{
					LetStatement{
						Literal: lexer.Token{
							Token:   lexer.LET,
							Literal: "let",
						},
						Identifier: IdentifierExpression{
							identifier: lexer.Token{
								Token:   lexer.IDENT,
								Literal: "x",
							},
						},
						Expression: IntegerExpression{
							token: lexer.Token{
								Token:   lexer.NUMBER,
								Literal: "10",
							},
							Val: 10,
						},
					},
					LetStatement{
						Literal: lexer.Token{
							Token:   lexer.LET,
							Literal: "let",
						},
						Identifier: IdentifierExpression{
							identifier: lexer.Token{
								Token:   lexer.IDENT,
								Literal: "b",
							},
						},
						Expression: IntegerExpression{
							token: lexer.Token{
								Token:   lexer.NUMBER,
								Literal: "5",
							},
							Val: 5,
						},
					},
					ExpressionStatement{
						Expr: IfExpression{
							token: lexer.Token{
								Token:   lexer.IF,
								Literal: "if",
							},
							condition: &InfixExpression{
								Left: IdentifierExpression{
									identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "a",
									},
								},
								Operator: lexer.Token{
									Token:   lexer.GT,
									Literal: ">",
								},
								Right: IdentifierExpression{
									identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "b",
									},
								},
							},
							consequence: BlockStatement{
								token: lexer.Token{
									Token:   lexer.BRLEFT,
									Literal: "{",
								},
								statements: []Statement{
									ExpressionStatement{
										Expr: IfExpression{
											token: lexer.Token{
												Token:   lexer.IF,
												Literal: "if",
											},
											condition: &InfixExpression{
												Left: IdentifierExpression{
													identifier: lexer.Token{
														Token:   lexer.IDENT,
														Line:    0,
														Column:  0,
														Literal: "b",
													},
												},
												Operator: lexer.Token{
													Token:   lexer.EQ,
													Literal: "==",
												},
												Right: IntegerExpression{
													token: lexer.Token{},
													Val:   5,
												},
											},
											consequence: BlockStatement{
												statements: []Statement{
													LetStatement{
														Literal: lexer.Token{
															Token:   lexer.LET,
															Literal: "let",
														},
														Identifier: IdentifierExpression{
															identifier: lexer.Token{
																Token:   lexer.IDENT,
																Literal: "b",
															},
														},
														Expression: IntegerExpression{
															token: lexer.Token{},
															Val:   10,
														},
													},
												},
											},
											alternative: nil,
										},
										Tok: lexer.Token{},
									},

									LetStatement{
										Literal: lexer.Token{
											Token:   lexer.LET,
											Literal: "let",
										},
										Identifier: IdentifierExpression{
											identifier: lexer.Token{
												Token:   lexer.IDENT,
												Literal: "a",
											},
										},
										Expression: IntegerExpression{
											token: lexer.Token{},
											Val:   10,
										},
									},
								},
							},
							alternative: &BlockStatement{
								statements: []Statement{
									LetStatement{
										Literal: lexer.Token{
											Token:   lexer.LET,
											Literal: "let",
										},
										Identifier: IdentifierExpression{
											identifier: lexer.Token{
												Token:   lexer.IDENT,
												Literal: "b",
											},
										},
										Expression: IntegerExpression{
											token: lexer.Token{
												Token:   lexer.NUMBER,
												Literal: "5",
											},
											Val: 5,
										},
									},
								},
							},
						},
						Tok: lexer.Token{
							Token:   lexer.IF,
							Literal: "if",
						},
					},
				},
			},
		},
		{
			i: `fn (a, b) {
				let a = 10;
			}`,
			o: &RootNode{
				Statements: []Statement{
					ExpressionStatement{
						Expr: FuncExpression{
							token: lexer.Token{
								Token:   lexer.FUNC,
								Literal: "fn",
							},
							args: []IdentifierExpression{
								{
									identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "a",
									},
								},
								{
									identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "b",
									},
								},
							},
							body: BlockStatement{
								token: lexer.Token{
									Token:   lexer.BRLEFT,
									Literal: "{",
								},
								statements: []Statement{
									LetStatement{
										Literal: lexer.Token{
											Token:   lexer.LET,
											Literal: "let",
										},
										Identifier: IdentifierExpression{
											identifier: lexer.Token{
												Token:   lexer.IDENT,
												Literal: "let",
											},
										},
										Expression: IntegerExpression{
											Val: 10,
										},
									},
								},
							},
						},
						Tok: lexer.Token{},
					},
				},
			},
		},
		{
			i: "a = 10;",
			o: &RootNode{
				Statements: []Statement{
					ExpressionStatement{
						Expr: AssignExpression{
							token: lexer.Token{
								Token:   lexer.ASSIGN,
								Literal: "=",
							},
							Identifier: IdentifierExpression{
								identifier: lexer.Token{
									Token:   lexer.IDENT,
									Literal: "a",
								},
							},
							Val: IntegerExpression{
								Val: 10,
							},
						},
						Tok: lexer.Token{
							Token:   lexer.IDENT,
							Literal: "a",
						},
					},
				},
			},
		},
		{
			i: `return (1 + 1);`,
			o: &RootNode{
				Statements: []Statement{
					ReturnStatement{
						token: lexer.Token{
							Token:   lexer.RETURN,
							Literal: "return",
						},
						returnExpr: &InfixExpression{
							Left: IntegerExpression{
								Val: 1,
							},
							Operator: lexer.Token{
								Token:   lexer.PLUS,
								Literal: "+",
							},
							Right: IntegerExpression{
								Val: 1,
							},
						},
					},
				},
			},
		},
	}

	for i, test := range ts {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			lex := lexer.New(bytes.NewBuffer([]byte(test.i)))
			if _, err := lex.Advance(); err != nil {
				t.Error(err)
			}
			p := NewParserFromLexer(lex)
			root, err := p.Parse()
			if err != nil {
				t.Error(err)
			}

			if len(p.Errors) != 0 {
				t.Error(p.Errors)
			}

			if !AssertNodes(t, root, test.o.(*RootNode)) {
				t.Errorf("Assertion failed, expected %q, got %q", test.o, root)
			}
		})

	}

}
