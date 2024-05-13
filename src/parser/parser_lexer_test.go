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
	t.Logf("Asserting %T and %T, %q, %q\n", a, b, a, b)
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
		if v.Identifier.Literal != b.(IdentifierExpression).Identifier.Literal {
			return false
		}
	case LetStatement:
		if !AssertNodes(t, v.Expression, b.(LetStatement).Expression) {
			return false
		}
	case *BlockStatement:
		if len(v.Statements) != len(b.(*BlockStatement).Statements) {
			t.Errorf("Length of Blockstatement is not equal: %d, %d\n", len(v.Statements), len(b.(*BlockStatement).Statements))
			return false
		}

		for i, _ := range v.Statements {
			if !AssertNodes(t, v.Statements[i], b.(*BlockStatement).Statements[i]) {
				return false
			}
		}
	case BlockStatement:
		if len(v.Statements) != len(b.(BlockStatement).Statements) {
			t.Errorf("Length of Blockstatement is not equal: %d, %d\n", len(v.Statements), len(b.(BlockStatement).Statements))
			return false
		}

		for i, _ := range v.Statements {
			if !AssertNodes(t, v.Statements[i], b.(BlockStatement).Statements[i]) {
				return false
			}
		}
	case IfExpression:
		if !AssertNodes(t, v.Condition, b.(IfExpression).Condition) || !AssertNodes(t, v.Consequence, b.(IfExpression).Consequence) {
			return false
		}

		if v.Alternative != nil && !AssertNodes(t, v.Alternative, b.(IfExpression).Alternative) {
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
		if len(v.Args) != len(b.(FuncExpression).Args) {
			t.Errorf("number of argumnets does not match\n")
			return false
		}
		for i, _ := range v.Args {
			if !AssertNodes(t, v.Args[i], b.(FuncExpression).Args[i]) {
				return false
			}
		}

		if len(v.Body.Statements) != len(b.(FuncExpression).Body.Statements) {
			t.Errorf("number of Statements in Body does not match\n")
			return false
		}

		for i, _ := range v.Body.Statements {
			return AssertNodes(t, v.Body.Statements[i], b.(FuncExpression).Body.Statements[i])
		}
	case AssignExpression:
		if !AssertNodes(t, v.Identifier, b.(AssignExpression).Identifier) {
			return false
		}

		if !AssertNodes(t, v.Val, b.(AssignExpression).Val) {
			return false
		}
	case ReturnStatement:
		if !AssertNodes(t, v.ReturnExpr, b.(ReturnStatement).ReturnExpr) {
			return false
		}

		if v.token.Token != b.(ReturnStatement).token.Token && v.token.Literal != b.(ReturnStatement).token.Literal {
			t.Errorf("return statement tokens are different")
			return false
		}
	case CallExpression:
		if !AssertNodes(t, v.Call, b.(CallExpression).Call) {
			t.Errorf("failed to assert call expression: %q, %q\n", v.Call, b.(CallExpression).Call)
			return false
		}

		if len(v.CallArgs) != len(b.(CallExpression).CallArgs) {
			t.Errorf("failed to assert number of call argumnets, %d, %d\n", len(v.CallArgs), len(b.(CallExpression).CallArgs))
			return false
		}

		for i, _ := range v.CallArgs {
			if !AssertNodes(t, v.CallArgs[i], b.(CallExpression).CallArgs[i]) {
				t.Errorf("failed to assert call args: %q, %q\n", v.CallArgs[i], b.(CallExpression).CallArgs[i])
				return false
			}
		}
	case BoolExpression:
		if v.Val != b.(BoolExpression).Val {
			t.Errorf("failed to assert bools: %v, %v\n", v.Val, b.(BoolExpression).Val)
			return false
		}
	case StringExpression:
		if v.Val != b.(StringExpression).Val {
			t.Errorf("failed to assert strings: %q, %q\n", v.Val, b.(StringExpression).Val)
			return false
		}
	case ArrayExpression:
		if len(v.Arr) != len(b.(ArrayExpression).Arr) {
			t.Errorf("failed to assert array lenghtes\n")
			return false
		}

		for i := range v.Arr {
			if !AssertNodes(t, v.Arr[i], b.(ArrayExpression).Arr[i]) {
				return false
			}
		}
	case HashMapExpression:
		if len(v.Map) != len(b.(HashMapExpression).Map) {
			t.Errorf("failed to assert map lengthes\n")
		}
	case NilExpression:
		return true
	case IndexExpression:
		if !AssertNodes(t, v.Idx, b.(IndexExpression).Idx) {
			return false
		}

		if !AssertNodes(t, v.Of, b.(IndexExpression).Of) {
			return false
		}
	default:
		t.Errorf("Not supported type %T\n", v)
		return false
	}

	return true
}

func ParseRoot(t *testing.T, in []byte) *RootNode {
	p := NewParser(bytes.NewBuffer(in))
	root, err := p.Parse()
	if err != nil {
		t.Errorf("failed to parse %s\n", err)
	}

	if len(p.Errors) != 0 {
		t.Error(p.Errors)
	}

	return root.(*RootNode)
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
							Token:   lexer.OR,
							Literal: "||",
						},
						Right: IntegerExpression{
							Val: 2,
						},
					},
					Operator: lexer.Token{
						Token:   lexer.OR,
						Literal: "||",
					},
					Right: IntegerExpression{
						Val: 2,
					},
				},
				Operator: lexer.Token{
					Token:   lexer.AND,
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
		root := ParseRoot(t, []byte(test.i))
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
							Identifier: lexer.Token{
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
							Identifier: lexer.Token{
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
							Identifier: lexer.Token{
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
							Identifier: lexer.Token{
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
							Condition: &InfixExpression{
								Left: IdentifierExpression{
									Identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "a",
									},
								},
								Operator: lexer.Token{
									Token:   lexer.GT,
									Literal: ">",
								},
								Right: IdentifierExpression{
									Identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "b",
									},
								},
							},
							Consequence: BlockStatement{
								token: lexer.Token{
									Token:   lexer.BRLEFT,
									Literal: "{",
								},
								Statements: []Statement{
									ExpressionStatement{
										Expr: IfExpression{
											token: lexer.Token{
												Token:   lexer.IF,
												Literal: "if",
											},
											Condition: &InfixExpression{
												Left: IdentifierExpression{
													Identifier: lexer.Token{
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
											Consequence: BlockStatement{
												Statements: []Statement{
													LetStatement{
														Literal: lexer.Token{
															Token:   lexer.LET,
															Literal: "let",
														},
														Identifier: IdentifierExpression{
															Identifier: lexer.Token{
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
											Alternative: nil,
										},
										Tok: lexer.Token{},
									},

									LetStatement{
										Literal: lexer.Token{
											Token:   lexer.LET,
											Literal: "let",
										},
										Identifier: IdentifierExpression{
											Identifier: lexer.Token{
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
							Alternative: &BlockStatement{
								Statements: []Statement{
									LetStatement{
										Literal: lexer.Token{
											Token:   lexer.LET,
											Literal: "let",
										},
										Identifier: IdentifierExpression{
											Identifier: lexer.Token{
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
							Args: []IdentifierExpression{
								{
									Identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "a",
									},
								},
								{
									Identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "b",
									},
								},
							},
							Body: BlockStatement{
								token: lexer.Token{
									Token:   lexer.BRLEFT,
									Literal: "{",
								},
								Statements: []Statement{
									LetStatement{
										Literal: lexer.Token{
											Token:   lexer.LET,
											Literal: "let",
										},
										Identifier: IdentifierExpression{
											Identifier: lexer.Token{
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
								Identifier: lexer.Token{
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
						ReturnExpr: &InfixExpression{
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
		{
			i: "let a = 10;a return a;",
			o: &RootNode{
				Statements: []Statement{
					LetStatement{
						Literal: lexer.Token{
							Token:   lexer.LET,
							Literal: "let",
						},
						Identifier: IdentifierExpression{
							Identifier: lexer.Token{
								Token:   lexer.IDENT,
								Literal: "a",
							},
						},
						Expression: IntegerExpression{
							Val: 10,
						},
					},
					ExpressionStatement{
						Expr: IdentifierExpression{
							Identifier: lexer.Token{
								Token:   lexer.IDENT,
								Literal: "a",
							},
						},
					},
					ReturnStatement{
						token: lexer.Token{
							Token:   lexer.RETURN,
							Line:    0,
							Column:  0,
							Literal: "return",
						},
						ReturnExpr: IdentifierExpression{
							Identifier: lexer.Token{
								Token:   lexer.IDENT,
								Literal: "a",
							},
						},
					},
				},
			},
		},
		{
			i: "let a = 10\nfn (a,b){}(a,b)",
			o: &RootNode{
				Statements: []Statement{
					LetStatement{
						Literal: lexer.Token{
							Token:   lexer.LET,
							Literal: "let",
						},
						Identifier: IdentifierExpression{
							Identifier: lexer.Token{
								Token:   lexer.IDENT,
								Literal: "a",
							},
						},
						Expression: IntegerExpression{
							Val: 10,
						},
					},
					ExpressionStatement{
						Expr: CallExpression{
							token: lexer.Token{},
							Call: FuncExpression{
								token: lexer.Token{},
								Args: []IdentifierExpression{
									{
										Identifier: lexer.Token{
											Token:   lexer.IDENT,
											Literal: "a",
										},
									},
									{
										Identifier: lexer.Token{
											Token:   lexer.IDENT,
											Literal: "b",
										},
									},
								},
								Body: BlockStatement{},
							},
							CallArgs: []Expression{
								IdentifierExpression{
									Identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "a",
									},
								},
								IdentifierExpression{
									Identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "b",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			i: "a(a,b, c)",
			o: &RootNode{
				Statements: []Statement{
					ExpressionStatement{
						Expr: CallExpression{
							Call: IdentifierExpression{
								Identifier: lexer.Token{
									Token:   lexer.IDENT,
									Literal: "a",
								},
							},
							CallArgs: []Expression{
								IdentifierExpression{
									Identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "a",
									},
								},
								IdentifierExpression{
									Identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "b",
									},
								},
								IdentifierExpression{
									Identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "c",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"(true || false) == (true || true)",
			&RootNode{
				Statements: []Statement{
					ExpressionStatement{
						Expr: &InfixExpression{
							Left: &InfixExpression{
								Left: BoolExpression{
									token: lexer.Token{
										Token:   lexer.TRUE,
										Literal: "true",
									},
									Val: true,
								},
								Operator: lexer.Token{
									Token:   lexer.OR,
									Literal: "||",
								},
								Right: BoolExpression{
									token: lexer.Token{
										Token:   lexer.FALSE,
										Literal: "false",
									},
									Val: false,
								},
							},
							Operator: lexer.Token{
								Token:   lexer.EQ,
								Literal: "==",
							},
							Right: &InfixExpression{
								Left: BoolExpression{
									token: lexer.Token{
										Token:   lexer.TRUE,
										Literal: "true",
									},
									Val: true,
								},
								Operator: lexer.Token{
									Token:   lexer.OR,
									Literal: "||",
								},
								Right: BoolExpression{
									token: lexer.Token{
										Token:   lexer.TRUE,
										Literal: "true",
									},
									Val: true,
								},
							},
						},
					},
				},
			},
		},
		{
			i: `[1, [1, [2]], ["string"]]`,
			o: &RootNode{
				Statements: []Statement{
					ExpressionStatement{
						Expr: ArrayExpression{
							Arr: []Expression{
								IntegerExpression{
									token: lexer.Token{
										Token:   lexer.NUMBER,
										Literal: "1",
									},
									Val: 1,
								},
								ArrayExpression{
									Arr: []Expression{
										IntegerExpression{
											token: lexer.Token{
												Token:   lexer.NUMBER,
												Literal: "1",
											},
											Val: 1,
										},
										ArrayExpression{
											Arr: []Expression{
												IntegerExpression{
													token: lexer.Token{
														Token:   lexer.NUMBER,
														Literal: "2",
													},
													Val: 2,
												},
											},
											token: lexer.Token{
												Token:   lexer.SBLEFT,
												Literal: "[",
											},
										},
									},
									token: lexer.Token{
										Token:   lexer.SBLEFT,
										Literal: "[",
									},
								},
								ArrayExpression{
									Arr: []Expression{
										StringExpression{
											tok: lexer.Token{
												Token:   lexer.STRING,
												Literal: "string",
											},
											Val: "string",
										},
									},
									token: lexer.Token{
										Token:   lexer.SBLEFT,
										Literal: "[",
									},
								},
							},
							token: lexer.Token{
								Token:   lexer.SBLEFT,
								Literal: "[",
							},
						},
					},
				},
			},
		},
		{
			`let a={}`,
			&RootNode{
				Statements: []Statement{
					LetStatement{
						Literal: lexer.Token{
							Token:   lexer.LET,
							Literal: "let",
						},
						Identifier: IdentifierExpression{
							Identifier: lexer.Token{
								Token:   lexer.IDENT,
								Literal: "a",
							},
						},
						Expression: HashMapExpression{
							Map: make(map[Expression]Expression),
						},
					},
				},
			},
		},
		{
			`let a = { "a" : "b", true: {}, 1 : 1}`,
			&RootNode{
				Statements: []Statement{
					LetStatement{
						Literal: lexer.Token{
							Token:   lexer.LET,
							Literal: "let",
						},
						Identifier: IdentifierExpression{
							Identifier: lexer.Token{
								Token:   lexer.IDENT,
								Literal: "a",
							},
						},
						Expression: HashMapExpression{
							Map: map[Expression]Expression{
								StringExpression{
									Val: "a",
								}: StringExpression{
									Val: "b",
								},
								BoolExpression{
									Val: true,
								}: HashMapExpression{
									Map: map[Expression]Expression{},
								},
								IntegerExpression{
									Val: 1,
								}: IntegerExpression{
									Val: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			`[[]]`,
			&RootNode{
				Statements: []Statement{
					ExpressionStatement{
						Expr: ArrayExpression{
							Arr: []Expression{
								ArrayExpression{
									Arr: []Expression{},
								},
							},
						},
					},
				},
			},
		},
		{
			`a["a"] = "b"`,
			&RootNode{
				Statements: []Statement{
					ExpressionStatement{
						Expr: AssignExpression{
							Identifier: IndexExpression{
								Of: IdentifierExpression{
									Identifier: lexer.Token{
										Token:   lexer.IDENT,
										Literal: "a",
									},
								},
								Idx: StringExpression{
									Val: "a",
								},
							},
							Val: StringExpression{
								Val: "b",
							},
						},
					},
				},
			},
		},
		{
			`{ return }`,
			&RootNode{
				Statements: []Statement{
					BlockStatement{
						Statements: []Statement{
							ReturnStatement{
								token: lexer.Token{
									Token:   lexer.RETURN,
									Literal: "return",
								},
								ReturnExpr: NilExpression{},
							},
						},
					},
				},
			},
		},
		{
			``
		},
	}

	for i, test := range ts {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			root := ParseRoot(t, []byte(test.i))

			if !AssertNodes(t, root, test.o.(*RootNode)) {
				t.Errorf("Assertion failed, expected %q, got %q", test.o, root)
			}
		})

	}

}

func TestHashMap(t *testing.T) {
	in := `let a = { "a" : "b", true: false, 1 : 1}`

	root := ParseRoot(t, []byte(in))
	if len(root.Statements) != 1 {
		t.Errorf("expected 1 statement, got %d\n", len(root.Statements))
	}

	expr := AssertLetStatement(t, root.Statements[0])
	_, ok := expr.Expression.(HashMapExpression)
	if !ok {
		t.Errorf("expected map, got %T\n", expr.Expression)
	}

}
