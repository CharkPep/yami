package eval

import (
	"fmt"
	"github.com/charkpep/yad/src/object"
	"github.com/charkpep/yad/src/parser"
)

var (
	TRUE = object.BoolObject{
		OType: object.BOOL_OBJ,
		Val:   true,
	}
	FALSE = object.BoolObject{
		OType: object.BOOL_OBJ,
		Val:   false,
	}
)

type RuntimeError struct {
	msg  string
	node parser.Node
}

func (re RuntimeError) Error() string {
	if re.node != nil {
		return fmt.Sprintf("%s | %s", re.msg, re.node)
	}

	return fmt.Sprintf("%s | nil", re.msg)
}

func NewRuntimeError(msg string, obj parser.Node) RuntimeError {
	return RuntimeError{
		msg:  msg,
		node: obj,
	}
}

type Evaluator struct {
	env *object.Environment
}

func NewEvaluator() *Evaluator {
	return &Evaluator{
		env: object.NewEnv(),
	}
}

func (e Evaluator) Eval(node parser.Node) (object.Object, error) {
	switch v := node.(type) {
	case *parser.RootNode:
		return e.evalStatements(v.Statements)
	case parser.ExpressionStatement:
		return e.Eval(v.Expr)
	case parser.LetStatement:
		if _, ok := e.env.Get(v.Identifier.Token().Literal); ok {
			return nil, NewRuntimeError("identifier is already defined", v)
		}

		val, err := e.Eval(v.Expression)
		if err != nil {
			return nil, err
		}

		e.env.Set(v.Identifier.Token().Literal, val)
		return val, nil
	case parser.IntegerExpression:
		return object.IntegerObject{
			OType: object.INTEGER_OBJ,
			Val:   v.Val,
		}, nil
	case parser.AssignExpression:
		if _, ok := e.env.Get(v.Identifier.Token().Literal); !ok {
			return nil, NewRuntimeError("identifier is not defined", v)
		}

		val, err := e.Eval(v.Val)
		if err != nil {
			return nil, err
		}

		e.env.Set(v.Identifier.Token().Literal, val)
		return val, nil
	case parser.IdentifierExpression:
		val, ok := e.env.Get(v.Token().Literal)
		if !ok {
			return nil, NewRuntimeError("identifier is not defined", v)
		}

		return val, nil
	case parser.PrefixExpression:
		return e.evalPrefix(v)
	case parser.BoolExpression:
		return e.nativeBoolToObj(v.Val), nil
	case *parser.InfixExpression:
		return e.evalInfix(v)
	default:
		return nil, fmt.Errorf("unsupported node %T\n", v)
	}
}

func (e Evaluator) evalStatements(stmts []parser.Statement) (object.Object, error) {
	var res object.Object

	for _, stmt := range stmts {
		var err error
		res, err = e.Eval(stmt)
		if err != nil {
			return nil, err
		}

	}

	return res, nil
}

func (e Evaluator) evalInfix(infix *parser.InfixExpression) (object.Object, error) {
	left, err := e.Eval(infix.Left)
	if err != nil {
		return nil, err
	}

	right, err := e.Eval(infix.Right)
	if err != nil {
		return nil, err
	}

	switch {
	case right.Type() == object.INTEGER_OBJ && left.Type() == object.INTEGER_OBJ:
		return e.evalInfixInteger(infix, left.(object.IntegerObject), right.(object.IntegerObject))
	case right.Type() == object.BOOL_OBJ && left.Type() == object.BOOL_OBJ:
		return e.evalBoolInfix(infix, left.(object.BoolObject), right.(object.BoolObject))
	case right.Type() == object.BOOL_OBJ && left.Type() == object.INTEGER_OBJ:
		rightInt := e.boolObjToInt(right.(object.BoolObject))
		return e.evalInfixInteger(infix, left.(object.IntegerObject), rightInt)
	case right.Type() == object.INTEGER_OBJ && left.Type() == object.BOOL_OBJ:
		leftInt := e.boolObjToInt(left.(object.BoolObject))
		return e.evalInfixInteger(infix, leftInt, right.(object.IntegerObject))

	}
	return nil, nil
}

func (e Evaluator) evalInfixInteger(infix *parser.InfixExpression, left, right object.IntegerObject) (object.Object, error) {
	fmt.Println("Operator", infix.Operator.Literal)
	switch infix.Operator.Literal {
	case "+":
		return object.IntegerObject{
			OType: object.INTEGER_OBJ,
			Val:   left.Val + right.Val,
		}, nil
	case "-":
		return object.IntegerObject{
			OType: object.INTEGER_OBJ,
			Val:   left.Val - right.Val,
		}, nil
	case "*":
		return object.IntegerObject{
			OType: object.INTEGER_OBJ,
			Val:   left.Val * right.Val,
		}, nil
	case "/":
		if right.Val == 0 {
			return nil, NewRuntimeError("zero division", infix.Right)
		}

		return object.IntegerObject{
			OType: object.INTEGER_OBJ,
			Val:   left.Val / right.Val,
		}, nil
	case "==":
		return e.nativeBoolToObj(left.Val == right.Val), nil
	case "!=":
		return e.nativeBoolToObj(left.Val != right.Val), nil
	case ">":
		return e.nativeBoolToObj(left.Val > right.Val), nil
	case "<":
		return e.nativeBoolToObj(left.Val < right.Val), nil
	case ">=":
		return e.nativeBoolToObj(left.Val >= right.Val), nil
	case "<=":
		return e.nativeBoolToObj(left.Val <= right.Val), nil
	case "&":
		return object.IntegerObject{
			OType: object.INTEGER_OBJ,
			Val:   left.Val & right.Val,
		}, nil
	case "|":
		return object.IntegerObject{
			OType: object.INTEGER_OBJ,
			Val:   left.Val | right.Val,
		}, nil
	default:
		return nil, NewRuntimeError("operator is not supported for int types", infix)
	}
}

func (e Evaluator) evalBoolInfix(infix *parser.InfixExpression, left, right object.BoolObject) (object.Object, error) {
	switch infix.Operator.Literal {
	case "==":
		return e.nativeBoolToObj(left.Val == right.Val), nil
	case "!=":
		return e.nativeBoolToObj(left.Val != right.Val), nil
	case "&&":
		return e.nativeBoolToObj(left.Val && right.Val), nil
	case "||":
		return e.nativeBoolToObj(left.Val || right.Val), nil
	}

	return nil, NewRuntimeError("unexpected operator", infix)
}

func (e Evaluator) evalPrefix(node parser.PrefixExpression) (object.Object, error) {
	right, err := e.Eval(node.Expr)
	if err != nil {
		return nil, err
	}
	switch node.Prefix.Literal {
	case "!":
		res, err := e.evalObjToBool(right)
		if err != nil {
			return nil, NewRuntimeError(err.Error(), node)
		}
		res.Val = !res.Val
		return res, nil
	case "-":
		res, err := e.evalMinusPrefix(right)
		if err != nil {
			return nil, NewRuntimeError(err.Error(), node)
		}

		return res, nil
	}

	return nil, NewRuntimeError("unexpected prefix operator", node)
}

func (e Evaluator) evalMinusPrefix(obj object.Object) (object.Object, error) {
	switch v := obj.(type) {
	case object.IntegerObject:
		v.Val = -v.Val
		return v, nil
	}

	return nil, fmt.Errorf("unexpected object")
}

func (e Evaluator) evalObjToBool(obj object.Object) (object.BoolObject, error) {
	switch v := obj.(type) {
	case object.BoolObject:
		return v, nil
	case object.IntegerObject:
		if v.Val >= 1 {
			return TRUE, nil
		}
		return FALSE, nil
	default:
		return FALSE, fmt.Errorf("unexpected node")
	}
}

func (e Evaluator) boolObjToInt(obj object.BoolObject) object.IntegerObject {
	if obj.Val {
		return object.IntegerObject{
			OType: object.INTEGER_OBJ,
			Val:   1,
		}
	}

	return object.IntegerObject{
		OType: object.INTEGER_OBJ,
		Val:   0,
	}
}

func (e Evaluator) nativeBoolToObj(val bool) object.BoolObject {
	if val {
		return TRUE
	}

	return FALSE
}
