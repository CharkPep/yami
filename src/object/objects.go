package object

import (
	"bytes"
	"fmt"
	"github.com/charkpep/yad/src/parser"
	"strings"
)

var (
	FUNC_OBJ    ObjectType = "FUNC"
	RETURN_OBJ  ObjectType = "RETURN"
	INTEGER_OBJ ObjectType = "INTEGER"
	BOOL_OBJ    ObjectType = "BOOL"
	STRING_OBJ  ObjectType = "STRING"
	NIL_OBJ     ObjectType = "NIL"
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

func NewFuncObject(args []parser.IdentifierExpression, body parser.BlockStatement) FuncObject {
	return FuncObject{
		Args: args,
		Body: body,
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
