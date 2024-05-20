package object

import (
	"bytes"
	"fmt"
	"github.com/charkpep/yami/src/parser"
	"io"
	"os"
	"strings"
)

var (
	FUNC_OBJ    ObjectType = "FUNC"
	RETURN_OBJ  ObjectType = "RETURN"
	INTEGER_OBJ ObjectType = "INTEGER"
	BOOL_OBJ    ObjectType = "BOOL"
	STRING_OBJ  ObjectType = "STRING"
	NIL_OBJ     ObjectType = "NIL"
	ARRAY_OBJ   ObjectType = "ARRAY"
	MAP_OBJ     ObjectType = "MAP"
	BUILDIN_OBJ ObjectType = "BUILDIN"
)

var (
	TRUE = BoolObject{
		Val: true,
	}
	FALSE = BoolObject{
		Val: false,
	}
	NIL = NilObject{}
)

type IntegerObject struct {
	Val int64
}

func (i IntegerObject) Type() ObjectType {
	return INTEGER_OBJ
}

func (i IntegerObject) Inspect() string {
	return fmt.Sprint(i.Val)
}

type BoolObject struct {
	Val bool
}

func (b BoolObject) Type() ObjectType {
	return BOOL_OBJ
}

func (b BoolObject) Inspect() string {
	return fmt.Sprint(b.Val)
}

type NilObject struct{}

func (n NilObject) Inspect() string {
	return "nil"
}

func (n NilObject) Type() ObjectType {
	return NIL_OBJ
}

type ReturnObject struct {
	Val Object
}

func (r ReturnObject) Type() ObjectType {
	return RETURN_OBJ
}

func (r ReturnObject) Inspect() string {
	return r.Val.Inspect()
}

type FuncObject struct {
	Args []parser.IdentifierExpression
	Body parser.BlockStatement
	Env  *Environment
}

func (f FuncObject) Type() ObjectType {
	return FUNC_OBJ
}

func (f FuncObject) Inspect() string {
	var buff bytes.Buffer

	buff.WriteString("fn (")
	params := make([]string, 0, len(f.Args))
	for _, arg := range f.Args {
		params = append(params, arg.String())
	}

	buff.WriteString(strings.Join(params, ","))
	buff.WriteString(")")
	buff.WriteString(f.Body.String())
	return buff.String()
}

func NewFuncObject(args []parser.IdentifierExpression, body parser.BlockStatement, env *Environment) FuncObject {
	return FuncObject{
		Args: args,
		Body: body,
		Env:  env,
	}
}

type StringObject struct {
	Val string
}

func (str StringObject) Type() ObjectType {
	return STRING_OBJ
}

func (str StringObject) Inspect() string {
	return str.Val
}

type ArrayObject struct {
	Val []Object
}

func (arr *ArrayObject) Type() ObjectType {
	return ARRAY_OBJ
}

func (arr *ArrayObject) Inspect() string {
	var buff bytes.Buffer
	buff.WriteString("[")
	objects := make([]string, 0, len(arr.Val))
	for _, arg := range arr.Val {
		objects = append(objects, arg.Inspect())
	}
	buff.WriteString(strings.Join(objects, ","))
	buff.WriteString("]")
	return buff.String()
}

type MapObject struct {
	Val map[Object]Object
}

func (mp MapObject) Type() ObjectType {
	return MAP_OBJ
}

func (mp MapObject) Inspect() string {
	var buff bytes.Buffer
	buff.WriteString("{")
	var elements []string
	for k, v := range mp.Val {
		elements = append(elements, fmt.Sprintf("%s:%s", k.Inspect(), v.Inspect()))
	}

	buff.WriteString(strings.Join(elements, ","))
	buff.WriteString("}")
	return buff.String()
}

// BuildInFunc nodes like len, print
type BuildInFunc func(args ...Object) (Object, error)

func (b BuildInFunc) Inspect() string {
	return "build in"
}

func (b BuildInFunc) Type() ObjectType {
	return BUILDIN_OBJ
}

var BuildIns = map[string]BuildInFunc{
	"len": func(args ...Object) (Object, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
		}

		switch v := args[0].(type) {
		case StringObject:
			return IntegerObject{Val: int64(len(v.Val))}, nil
		case *ArrayObject:
			return IntegerObject{Val: int64(len(v.Val))}, nil
		default:
			return nil, fmt.Errorf("unexpected argument type")
		}
	},
	"print": func(args ...Object) (Object, error) {
		for _, arg := range args {
			io.WriteString(os.Stdout, arg.Inspect())
			io.WriteString(os.Stdout, "\n")
		}

		return NIL, nil
	},
}
