package eval

import (
	"bytes"
	"fmt"
	"github.com/charkpep/yad/src/object"
	"github.com/charkpep/yad/src/parser"
	"io"
	"reflect"
	"testing"
)

func CreateEvaluator(t *testing.T, in io.Reader) object.Object {
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
		t.Logf("Got different value kinds: (%v, %s), (%v, %s)\n", a, reflect.ValueOf(a).Kind(), b, reflect.ValueOf(b).Kind())
		return false
	}

	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		t.Logf("Got different value types: (%v, %T), (%v, %T)\n", a, a, b, b)
		return false
	}

	switch v := a.(type) {
	case object.IntegerObject:
		return v.Val != b.(object.IntegerObject).Val
	default:
		t.Errorf("Not implemented for object %T\n", a)
		return false
	}

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
		// TODO add bitwise operators
		//{"4 & 12", "4"},
		//{"4 & 12", "12"},
	}

	for i, test := range ts {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			obj := CreateEvaluator(t, bytes.NewBufferString(test.i))
			if obj.Inspect() != test.e {
				t.Errorf("expected %s, got %s\n", test.e, obj.Inspect())
			}
		})
	}
}

func TestAssignment(t *testing.T) {
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
	}

	for i, test := range ts {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			obj := CreateEvaluator(t, bytes.NewBufferString(test.i))
			AssertObjects(t, obj, test.o)
		})
	}

}
