package object

import "fmt"

var (
	INTEGER_OBJ ObjectType = "INTEGER"
	BOOL_OBJ    ObjectType = "BOOL"
	NIL_OBJ     ObjectType = "NIL"
)

type IntegerObject struct {
	OType ObjectType
	Val   int64
}

func (i IntegerObject) Type() ObjectType {
	return i.OType
}

func (i IntegerObject) Inspect() string {
	return fmt.Sprint(i.Val)
}

type BoolObject struct {
	OType ObjectType
	Val   bool
}

func (b BoolObject) Type() ObjectType {
	return b.OType
}

func (b BoolObject) Inspect() string {
	return fmt.Sprint(b.Val)
}
