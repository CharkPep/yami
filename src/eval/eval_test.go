package eval

import (
	"bytes"
	"fmt"
	"github.com/charkpep/yad/src/lexer"
	"github.com/charkpep/yad/src/object"
	"github.com/charkpep/yad/src/parser"
	"io"
	"reflect"
	"testing"
)

func EvaluateProgram(t *testing.T, in io.Reader) object.Object {
	p := parser.NewParser(in)
	root, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Errors) != 0 {
		t.Fatal(p.Errors)
	}

	e := NewEvaluator()

	obj, err := e.Eval(root)
	if err != nil {
		t.Fatal(err)
	}

	return obj
}

func AssertObjects(t *testing.T, a, b object.Object) bool {
	t.Logf("Asserting %T and %T, %+v, %+v\n", a, b, a, b)
	if reflect.ValueOf(a).Kind() != reflect.ValueOf(b).Kind() {
		t.Errorf("Got different value kinds: (%v, %s), (%v, %s)\n", a, reflect.ValueOf(a).Kind(), b, reflect.ValueOf(b).Kind())
		return false
	}

	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		t.Errorf("Got different value types: (%v, %T), (%v, %T)\n", a, a, b, b)
		return false
	}

	switch v := a.(type) {
	case object.IntegerObject:
		if v.Val != b.(object.IntegerObject).Val {
			t.Errorf("failed to assert integer objects: %v, %v\n", v, b)
			return false
		}
	case object.BoolObject:
		if v.Val != b.(object.BoolObject).Val {
			t.Errorf("failed to assert bool objects: %v, %v\n", v, b)
			return false
		}
	case object.FuncObject:
		for i, param := range v.Args {
			if param.String() != b.(object.FuncObject).Args[i].String() {
				t.Errorf("failed to assert function identifiers, got: %s, %s\n", param, b.(object.FuncObject).Args[i])
				return false
			}
		}

		if v.Body.String() != b.(object.FuncObject).Body.String() {
			t.Errorf("failed to assert string representation of function body")
			return false
		}
	case object.NilObject:
		return true
	case object.StringObject:
		if v.Val != b.(object.StringObject).Val {
			t.Errorf("failed to assert string values")
			return false
		}
	case *object.ArrayObject:
		bArr := b.(*object.ArrayObject)
		if len(v.Val) != len(bArr.Val) {
			t.Errorf("failed to assert array lengthes")
			return false
		}

		for i := range v.Val {
			if !AssertObjects(t, v.Val[i], b.(*object.ArrayObject).Val[i]) {
				return false
			}
		}
	case object.MapObject:
		if len(v.Val) != len(b.(object.MapObject).Val) {
			t.Errorf("failed to assert map lengthes")
			return false
		}

		for k := range v.Val {
			if !AssertObjects(t, v.Val[k], b.(object.MapObject).Val[k]) {
				return false
			}
		}
	default:
		t.Errorf("Not implemented for object %T\n", a)
		return false
	}

	return true
}

func TestEvalArithmetic(t *testing.T) {
	type tt struct {
		i string
		e string
	}

	ts := []tt{
		{"1", "1"},
		{"true", "true"},
		{"false", "false"},
		{"!1", "false"},
		{"-1", "-1"},
		{"1 + 1 * 2 / 2", "2"},
		{"1 + (2 * 2) / 2", "3"},
		{"true && false", "false"},
		{"true || false == true || true", "true"},
		{"1 + true", "2"},
		{"1 - true", "0"},
		{"1 + false", "1"},
		{"10 + false", "10"},
		{`"hello " + "world"`, "hello world"},
		{`"" + "hello"`, `hello`},
		{`"hello"[0]`, "h"},
		{`"f"[0]`, "f"},
		{"4 & 12", "4"},
		{"4 | 12", "12"},
		{"1 << 16", fmt.Sprint(1 << 16)},
		{"256 >> 7", "2"},
		{"1 || 1", "true"},
		{"(256 >> 7 < 256 >> 6) || 256 << 7 ", "true"},
		{"(256 >> 7 < 256 >> 6) && 256 << 7 ", "true"},
	}

	for i, test := range ts {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			obj := EvaluateProgram(t, bytes.NewBufferString(test.i))
			if obj.Inspect() != test.e {
				t.Errorf("expected %q, got %q\n", test.e, obj.Inspect())
			}
		})
	}
}

func TestEval(t *testing.T) {
	type tt struct {
		i string
		o object.Object
	}

	ts := []tt{
		{
			i: "let a = 10;\na",
			o: object.IntegerObject{
				Val: 10,
			},
		},
		{
			i: "let a = 10;\na=20;\na",
			o: object.IntegerObject{
				Val: 20,
			},
		},
		{
			i: "let a = -10;\na",
			o: object.IntegerObject{
				Val: -10,
			},
		},
		{
			i: "let a = !true;\na",
			o: object.BoolObject{
				Val: false,
			},
		},
		{
			i: `if 10 == 10 {
					10
				}`,
			o: object.IntegerObject{
				Val: 10,
			},
		},
		{
			i: `let a = 10
				if a > 10 {
					a
				} else {
					a = 2000
					a
				}`,
			o: object.IntegerObject{
				Val: 2000,
			},
		},
		{
			"let a = 10;\nreturn a; a=20",
			object.IntegerObject{
				Val: 10,
			},
		},
		{
			"{\nlet a = 10}\nlet a = 5;\na",
			object.IntegerObject{
				Val: 5,
			},
		},
		{
			`{
					let a = 10
					return a
					a = 20
				}`,
			object.IntegerObject{
				Val: 10,
			},
		},
		{
			`fn (b, c) {}`,
			object.FuncObject{
				Args: []parser.IdentifierExpression{
					{
						Identifier: lexer.Token{
							Token:   lexer.IDENT,
							Literal: "b",
						},
					},
					{
						Identifier: lexer.Token{
							Token:   lexer.IDENT,
							Literal: "c",
						},
					},
				},
				Body: parser.BlockStatement{},
			},
		},
		{
			`let a = fn (b, c) {}`,
			object.FuncObject{
				Args: []parser.IdentifierExpression{
					{
						Identifier: lexer.Token{
							Token:   lexer.IDENT,
							Literal: "b",
						},
					},
					{
						Identifier: lexer.Token{
							Token:   lexer.IDENT,
							Literal: "c",
						},
					},
				},
				Body: parser.BlockStatement{},
			},
		},
		{
			`let b = 1
				let c = 2
			fn (b, c) {b}(b,c)`,
			object.IntegerObject{
				Val: 1,
			},
		},
		{
			`let b = 1
			let c = 2
			fn (b, c) {
				b
				return c
				b
			}(b,c)`,
			object.IntegerObject{
				Val: 2,
			},
		},
		{
			`let add = fn (a, num) {
				return a + num
			}
			return add(2, add(2, 10))	
			`,
			object.IntegerObject{
				Val: 14,
			},
		},
		{
			`let factor = fn(n) {
				if n == 1 { return 1 }
				return n*factor(n-1)
			}
			return factor(5)`,
			object.IntegerObject{
				Val: 120,
			},
		},
		{
			`if 1 == 1 {
				return 1
			 }
			return 2`,
			object.IntegerObject{
				Val: 1,
			},
		},
		{
			`let n = 10;
			let fib = fn (cur, prev, cur_n) {
					if cur_n == n {
						return cur
					}
					return fib(cur + prev, cur, cur_n + 1)
				}
			fib(0, 1, 0)
			`,
			object.IntegerObject{
				Val: 55,
			},
		},
		{
			`if 1==0 {return 0}`,
			object.NilObject{},
		},
		{
			`"hello"`,
			object.StringObject{
				Val: "hello",
			},
		},
		{
			`return "hello world"`,
			object.StringObject{
				Val: "hello world",
			},
		},
		{
			`return "hello"[0] + "ello"`,
			object.StringObject{
				Val: "hello",
			},
		},
		{
			`let a = "h"; a = a[0]; a`,
			object.StringObject{
				Val: "h",
			},
		},
		{
			`[1,2,"string", [1,2]]`,
			&object.ArrayObject{
				Val: []object.Object{
					object.IntegerObject{
						Val: 1,
					},
					object.IntegerObject{
						Val: 2,
					},
					object.StringObject{
						Val: "string",
					},
					&object.ArrayObject{
						Val: []object.Object{
							object.IntegerObject{
								Val: 1,
							},
							object.IntegerObject{
								Val: 2,
							},
						},
					},
				},
			},
		},
		{
			`[1,2,3][0]`,
			object.IntegerObject{
				Val: 1,
			},
		},
		{
			`[1,2,"string"][2]`,
			object.StringObject{
				Val: "string",
			},
		},
		{
			`let a = { "a": "b"}`,
			object.MapObject{
				Val: map[object.Object]object.Object{
					object.StringObject{
						Val: "a",
					}: object.StringObject{
						Val: "b",
					},
				},
			},
		},
		{
			`let a = { "a": "b"}; a["a"]`,
			object.StringObject{
				Val: "b",
			},
		},
		{
			`let a = { true: false}; a[true]`,
			object.BoolObject{
				Val: false,
			},
		},
		{
			`len("")`,
			object.IntegerObject{
				Val: 0,
			},
		},
		{
			`len("hello")`,
			object.IntegerObject{
				Val: 5,
			},
		},
		{
			`let a = {}; if a[1] { "bad" } else { "pass" }`,
			object.StringObject{
				Val: "pass",
			},
		},
		{
			`let count = fn() { let counter = 0; return fn() { counter = counter + 1; return counter } } let c = count(); c()`,
			object.IntegerObject{
				Val: 1,
			},
		},
		{
			`let n = 10; let set = fn(n) { n }; set(1)`,
			object.IntegerObject{Val: 1},
		},
		{
			`{}`,
			object.NIL,
		},
		{
			`let a = fn() {
				let n = 10
				let b = fn() {
					n = n + 1
				}
				b()
				return n
			}
			a()`,
			object.IntegerObject{
				Val: 11,
			},
		},
		{
			`let a = [[1]]; a[0][0]`,
			object.IntegerObject{
				Val: 1,
			},
		},
		{
			`let a = {}; a["a"] = "b"; a["a"]`,
			object.StringObject{
				Val: "b",
			},
		},
		{
			`let a = [1]; a[0] = 10; a[0]`,
			object.IntegerObject{
				Val: 10,
			},
		},
		{
			`let s = "str"; s[0] = "a"; s`,
			object.StringObject{
				Val: "atr",
			},
		},
	}

	for i, test := range ts {
		t.Log(test.i)
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			obj := EvaluateProgram(t, bytes.NewBufferString(test.i))
			AssertObjects(t, obj, test.o)
		})
	}

}
