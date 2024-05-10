package eval

import (
	"fmt"
	"github.com/charkpep/yad/src/object"
	"github.com/charkpep/yad/src/parser"
)

var (
	TRUE = object.BoolObject{
		Val: true,
	}
	FALSE = object.BoolObject{
		Val: false,
	}
	NIL = object.NilObject{}
)

type RuntimeError struct {
	msg  string
	node parser.Node
}

func (re RuntimeError) Error() string {
	if re.node != nil {
		return fmt.Sprintf("%s | %s line %d, column %d", re.msg, re.node, re.node.Token().Line, re.node.Token().Column)
	}

	return fmt.Sprintf("%s | nil", re.msg)
}

func NewRuntimeError(msg string, obj parser.Node) RuntimeError {
	return RuntimeError{
		msg:  msg,
		node: obj,
	}
}

type Evaluator struct{}

func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

func (e Evaluator) Eval(node parser.Node) (object.Object, error) {
	env := object.NewEnv()
	return e.eval(node, env)
}

func (e Evaluator) EvalWithEnv(node parser.Node, env *object.Environment) (object.Object, error) {
	return e.eval(node, env)
}

func (e Evaluator) eval(node parser.Node, env *object.Environment) (object.Object, error) {
	switch v := node.(type) {
	case *parser.RootNode:
		return e.evalStatements(v.Statements, env)
	case parser.ExpressionStatement:
		return e.eval(v.Expr, env)
	case parser.LetStatement:
		if _, ok := env.Get(v.Identifier.Token().Literal); ok {
			return nil, NewRuntimeError("identifier is already defined", v)
		}

		val, err := e.eval(v.Expression, env)
		if err != nil {
			return nil, err
		}

		env.Set(v.Identifier.Token().Literal, val)
		return val, nil
	case parser.IntegerExpression:
		return object.IntegerObject{
			Val: v.Val,
		}, nil
	case parser.AssignExpression:
		if _, ok := env.Get(v.Identifier.Token().Literal); !ok {
			return nil, NewRuntimeError("identifier is not defined", v)
		}

		val, err := e.eval(v.Val, env)
		if err != nil {
			return nil, err
		}

		env.Set(v.Identifier.Token().Literal, val)
		return val, nil
	case parser.IfExpression:
		condition, err := e.eval(v.Condition, env)
		if err != nil {
			return nil, err
		}

		cond, err := e.evalObjToBool(condition)
		if err != nil {
			return nil, err
		}

		if cond.Val {
			return e.evalBlockStatement(v.Consequence, env)
		}

		if v.Alternative != nil {
			return e.evalBlockStatement(*v.Alternative, env)
		}

		return NIL, nil
	case parser.FuncExpression:
		return object.NewFuncObject(v.Args, v.Body), nil
	case parser.BlockStatement:
		derivedEvn := object.DeriveEnv(env)
		return e.evalBlockStatement(v, derivedEvn)
	case parser.ReturnStatement:
		returnObj, err := e.eval(v.ReturnExpr, env)
		if err != nil {
			return nil, err
		}

		return object.ReturnObject{
			Val: returnObj,
		}, nil
	case parser.IdentifierExpression:
		val, ok := env.Get(v.Token().Literal)
		if !ok {
			return nil, NewRuntimeError("identifier is not defined", v)
		}

		return val, nil
	case parser.PrefixExpression:
		return e.evalPrefix(v, env)
	case parser.BoolExpression:
		return e.nativeBoolToObj(v.Val), nil
	case *parser.InfixExpression:
		eval, err := e.evalInfix(v, env)
		return eval, err
	case parser.CallExpression:
		return e.evalCallExpression(v, env)
	case parser.StringExpression:
		return object.StringObject{
			Val: v.Val,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported node %T\n", v)
	}
}

func (e Evaluator) evalBlockStatement(stmt parser.BlockStatement, env *object.Environment) (object.Object, error) {
	var res object.Object
	for _, stmt := range stmt.Statements {
		var err error
		res, err = e.eval(stmt, env)
		if err != nil {
			return nil, err
		}

		if res.Type() == object.RETURN_OBJ {
			return res, nil
		}

	}

	return res, nil

}

func (e Evaluator) evalCallExpression(expr parser.CallExpression, env *object.Environment) (object.Object, error) {
	callObj, err := e.eval(expr.Call, env)
	if err != nil {
		return nil, err
	}

	call, ok := callObj.(object.FuncObject)
	if !ok {
		return nil, NewRuntimeError("expected function expression", expr)
	}

	if len(call.Args) != len(expr.CallArgs) {
		return nil, NewRuntimeError("mismatching number of arguments", expr)
	}

	derivedEvn := object.DeriveEnv(env)
	for i, arg := range expr.CallArgs {
		obj, err := e.eval(arg, env)
		if err != nil {
			return nil, err
		}

		derivedEvn.Set(call.Args[i].Identifier.Literal, obj)
	}

	return e.evalStatements(call.Body.Statements, derivedEvn)
}

func (e Evaluator) evalStatements(stmts []parser.Statement, env *object.Environment) (object.Object, error) {
	var res object.Object

	for _, stmt := range stmts {
		var err error
		res, err = e.eval(stmt, env)
		if err != nil {
			return nil, err
		}

		if res.Type() == object.RETURN_OBJ {
			return res.(object.ReturnObject).Val, nil
		}

	}

	return res, nil
}

func (e Evaluator) evalInfix(infix *parser.InfixExpression, env *object.Environment) (object.Object, error) {
	left, err := e.eval(infix.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := e.eval(infix.Right, env)
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
	case right.Type() == object.STRING_OBJ && left.Type() == object.STRING_OBJ:
		return object.StringObject{
			Val: left.(object.StringObject).Val + right.(object.StringObject).Val,
		}, nil

	}

	return nil, fmt.Errorf("not supported types")
}

func (e Evaluator) evalInfixInteger(infix *parser.InfixExpression, left, right object.IntegerObject) (object.Object, error) {
	switch infix.Operator.Literal {
	case "+":
		return object.IntegerObject{
			Val: left.Val + right.Val,
		}, nil
	case "-":
		return object.IntegerObject{
			Val: left.Val - right.Val,
		}, nil
	case "*":
		return object.IntegerObject{
			Val: left.Val * right.Val,
		}, nil
	case "/":
		if right.Val == 0 {
			return nil, NewRuntimeError("zero division", infix.Right)
		}

		return object.IntegerObject{
			Val: left.Val / right.Val,
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
			Val: left.Val & right.Val,
		}, nil
	case "|":
		return object.IntegerObject{
			Val: left.Val | right.Val,
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

func (e Evaluator) evalPrefix(node parser.PrefixExpression, env *object.Environment) (object.Object, error) {
	right, err := e.eval(node.Expr, env)
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
			Val: 1,
		}
	}

	return object.IntegerObject{
		Val: 0,
	}
}

func (e Evaluator) nativeBoolToObj(val bool) object.BoolObject {
	if val {
		return TRUE
	}

	return FALSE
}
