package eval

import (
	"fmt"
	"github.com/charkpep/yad/src/object"
	"github.com/charkpep/yad/src/parser"
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

// TODO decouple in separate functions shit pile of switch case
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
		switch identifierExpression := v.Identifier.(type) {
		case parser.IndexExpression:
			ident, ok := identifierExpression.Of.(parser.IdentifierExpression)
			if !ok {
				return nil, NewRuntimeError("expected identifier for indexed assignment", identifierExpression)
			}

			structure, ok := env.Get(ident.Identifier.Literal)
			if !ok {
				return nil, NewRuntimeError("identifier is not defined", identifierExpression)
			}

			idx, err := e.eval(identifierExpression.Idx, env)
			if err != nil {
				return nil, err
			}

			val, err := e.eval(v.Val, env)
			if err != nil {
				return nil, err
			}

			switch {
			case structure.Type() == object.MAP_OBJ:
				structure.(object.MapObject).Val[idx] = val
				//case structure.Type() ==  object.ARRAY_OBJ && val.Type() == object.INTEGER_OBJ:
				//	if val.(object.IntegerObject).Val >= int64(len(structure.(*object.ArrayObject).Val)) {
				//		return nil, NewRuntimeError("index out of bounds", ident)
				//	}

			}

			return val, nil
		case parser.IdentifierExpression:
			val, err := e.eval(v.Val, env)
			if err != nil {
				return nil, err
			}

			env.Set(identifierExpression.Token().Literal, val)
			return val, nil
		default:
			return nil, NewRuntimeError("unsupported assignment", node)
		}
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

		return object.NIL, nil
	case parser.FuncExpression:
		return object.NewFuncObject(v.Args, v.Body, object.DeriveEnv(env)), nil
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
		if ok {
			return val, nil
		}

		val, ok = object.BuildIns[v.Token().Literal]
		if ok {
			return val, nil
		}

		return nil, NewRuntimeError("identifier is not defined", v)
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
	case parser.IndexExpression:
		return e.evalIndex(v, env)
	case parser.ArrayExpression:
		return e.evalArray(v, env)
	case parser.HashMapExpression:
		return e.evalMap(v, env)
	default:
		return nil, fmt.Errorf("unsupported node %T\n", v)
	}
}

func (e Evaluator) evalMap(mp parser.HashMapExpression, env *object.Environment) (object.Object, error) {
	mpObj := object.MapObject{
		Val: make(map[object.Object]object.Object),
	}

	for k, v := range mp.Map {
		key, err := e.eval(k, env)
		if err != nil {
			return nil, err
		}
		val, err := e.eval(v, env)
		if err != nil {
			return nil, err
		}

		mpObj.Val[key] = val
	}

	return mpObj, nil
}

func (e Evaluator) evalArray(arrExpr parser.ArrayExpression, env *object.Environment) (object.Object, error) {
	arr := &object.ArrayObject{
		Val: make([]object.Object, 0, len(arrExpr.Arr)),
	}

	for _, expr := range arrExpr.Arr {
		obj, err := e.eval(expr, env)
		if err != nil {
			return nil, err
		}
		arr.Val = append(arr.Val, obj)
	}

	return arr, nil
}
func (e Evaluator) evalIndex(expr parser.IndexExpression, env *object.Environment) (object.Object, error) {
	ofObj, err := e.eval(expr.Of, env)
	if err != nil {
		return nil, err
	}

	idx, err := e.eval(expr.Idx, env)
	if err != nil {
		return nil, err
	}

	switch {
	case idx.Type() == object.INTEGER_OBJ && ofObj.Type() == object.STRING_OBJ:
		index := idx.(object.IntegerObject).Val
		str := ofObj.(object.StringObject).Val
		if index >= int64(len(str)) {
			return nil, NewRuntimeError("index out of bounds", expr)
		}

		return object.StringObject{
			Val: string(str[index]),
		}, nil
	case idx.Type() == object.INTEGER_OBJ && ofObj.Type() == object.ARRAY_OBJ:
		return ofObj.(*object.ArrayObject).Val[idx.(object.IntegerObject).Val], nil
	case ofObj.Type() == object.MAP_OBJ:
		val, ok := ofObj.(object.MapObject).Val[idx]
		if !ok {
			val = object.NIL
		}
		return val, nil
	default:
		return nil, NewRuntimeError("unexpected index type for expression", expr)
	}

}

func (e Evaluator) evalBlockStatement(stmt parser.BlockStatement, env *object.Environment) (object.Object, error) {
	var res object.Object = object.NIL
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

	switch call := callObj.(type) {
	case object.FuncObject:
		if len(call.Args) != len(expr.CallArgs) {
			return nil, NewRuntimeError("mismatching number of arguments", expr)
		}

		objs, err := e.evalExpressions(expr.CallArgs, env)
		if err != nil {
			return nil, err
		}

		for i, k := range call.Args {
			call.Env.Set(k.Identifier.Literal, objs[i])
		}

		return e.evalStatements(call.Body.Statements, call.Env)
	case object.BuildInFunc:
		objs, err := e.evalExpressions(expr.CallArgs, env)
		if err != nil {
			return nil, err
		}
		return call(objs...)
	default:
		return nil, NewRuntimeError("expected function expression", expr)
	}

}

func (e Evaluator) evalExpressions(args []parser.Expression, env *object.Environment) ([]object.Object, error) {
	objs := make([]object.Object, 0, len(args))
	for _, arg := range args {
		obj, err := e.eval(arg, env)
		if err != nil {
			return nil, err
		}

		objs = append(objs, obj)
	}

	return objs, nil

}

func (e Evaluator) evalStatements(stmts []parser.Statement, env *object.Environment) (object.Object, error) {
	var res object.Object = object.NIL
	for _, stmt := range stmts {
		var err error
		res, err = e.eval(stmt, env)
		if err != nil {
			return nil, err
		}

		if res.Type() != object.NIL_OBJ && res.Type() == object.RETURN_OBJ {
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

	return nil, NewRuntimeError("not supported types", infix)
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
	case "<<":
		return object.IntegerObject{
			Val: left.Val << right.Val,
		}, nil
	case ">>":
		return object.IntegerObject{
			Val: left.Val >> right.Val,
		}, nil
	case "||":
		return object.BoolObject{
			Val: left.Val > 0 || right.Val > 0,
		}, nil
	case "&&":
		return object.BoolObject{
			Val: left.Val > 0 && right.Val > 0,
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
	case object.NilObject:
		return object.FALSE, nil
	case object.BoolObject:
		return v, nil
	case object.IntegerObject:
		if v.Val >= 1 {
			return object.TRUE, nil
		}
		return object.FALSE, nil
	default:
		return object.FALSE, fmt.Errorf("unexpected node")

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
		return object.TRUE
	}

	return object.FALSE
}
