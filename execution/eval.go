package execution

import (
	"errors"
	"fmt"
	"interpreter/ast"
	"interpreter/value"
)

func call(env *Env, caller value.Object, args []value.Object) (value.Object, error) {
	switch caller := caller.(type) {
	case *value.Function:
		return callFunction(env, caller, args)
	case *value.NativeFunction:
		return caller.Call(args)
		// case *value.Class:
		// 	c := caller.(*value.Class)
		// 	f, err := c.Get("call")
		// 	if err != nil {
		// 		return nil, errors.New("class " + caller.Class() + " does not defined `call` method and is not callable.")
		// 	}
		// 	return call(env, f, args)
	}

	return nil, errors.New("value of type " + caller.Class() + " is not callable")
}

func callFunction(env *Env, f *value.Function, args []value.Object) (value.Object, error) {
	if len(args) != len(f.Args) {
		return nil, errors.New(fmt.Sprint("called function with", len(args), "arguments, expected", len(f.Args)))
	}

	env = env.NewScope()
	for i, ident := range f.Args {
		// after calling new scope the error cant be null
		_ = env.Set(ident, args[i])
	}

	return Eval(env, f.Body)
}

func Eval(env *Env, expr ast.Expression) (value.Object, error) {

	switch expr := expr.(type) {
	case *ast.ClassDefinition:
		err := defineClass(env, expr)
		if err != nil {
			return nil, err
		}
		return nil, nil

	case *ast.FunctionDefinition:
		err := defineFunc(env, expr)
		if err != nil {
			return nil, err
		}
		return nil, nil

	case *ast.DoFlow:
		var r value.Object
		var err error

		for _, statement := range expr.Statements {
			r, err = Eval(env, statement)
			if err != nil {
				return nil, err
			}
		}

		return r, nil

	case *ast.IfFlow:

		res, err := Eval(env, expr.Condition)
		if err != nil {
			return nil, err
		}

		if res.Boolean() {
			return Eval(env, expr.True)
		}
		return Eval(env, expr.False)

	case *ast.OrFlow:
		var r value.Object
		var err error

		for _, statement := range expr.Arguments {
			r, err = Eval(env, statement)
			if err != nil {
				return nil, err
			}

			if r.Boolean() {
				return r, nil
			}
		}

		return r, nil

	case *ast.AndFlow:
		var r value.Object
		var err error

		for _, statement := range expr.Arguments {
			r, err = Eval(env, statement)
			if err != nil {
				return nil, err
			}

			if !r.Boolean() {
				return r, nil
			}
		}

		return r, nil

	case *ast.NamedCall:

		function, err := env.Get(expr.Function)
		if err != nil {
			return nil, err
		}

		args := make([]value.Object, len(expr.Arguments))
		for _, arg := range expr.Arguments {
			value, err := Eval(env, arg)
			if err != nil {
				return nil, err
			}

			args = append(args, value)
		}

		return call(env, function, args)

	case *ast.Call:

		function, err := Eval(env, expr.Function)
		if err != nil {
			return nil, err
		}

		args := make([]value.Object, len(expr.Arguments))
		for _, arg := range expr.Arguments {
			value, err := Eval(env, arg)
			if err != nil {
				return nil, err
			}

			args = append(args, value)
		}

		return call(env, function, args)

	// Literals
	case *ast.IdentLiteral:
		return env.Get(expr.Value)

	case *ast.BoolLiteral:
		return value.NewBool(expr.Value), nil

	case *ast.NilLiteral:
		return value.Nil(), nil

	case *ast.IntLiteral:
		return value.NewInt(expr.Value), nil

	case *ast.FloatLiteral:
		return value.NewFloat(expr.Value), nil

	case *ast.StringLiteral:
		return value.NewString(expr.Value), nil

	case *ast.LambdaLiteral:
		return value.NewFunction(expr.Arguments, expr.Body)

	case *ast.ArrayLiteral:
		values := make([]value.Object, len(expr.Values))
		for _, expr := range expr.Values {
			v, err := Eval(env, expr)
			if err != nil {
				return nil, err
			}

			values = append(values, v)
		}
		return value.NewArray(values...), nil

	// (let x y body)
	case *ast.VariableDefiniton:
		// first evaluate variable assignment
		val, err := Eval(env, expr.Value)
		if err != nil {
			return nil, err
		}

		// overshadow what has been
		return env.LetIn(expr.Ident, val, func(env *Env) (value.Object, error) {
			return Eval(env, expr.Body)
		})
	}

	return nil, errors.New("unknown expression encountered")
}

func defineClass(env *Env, def *ast.ClassDefinition) error {
	classInfo, err := value.NewClassInfo(def.Name, def.Fields)
	if err != nil {
		return err
	}

	constructor := value.NewNativeFunction(classInfo.MakeInstance)

	env.DefineGlobal(def.Name, constructor)
	return nil
}

func defineFunc(env *Env, def *ast.FunctionDefinition) error {
	// TODO this is the spot to include information about codeposition in file, row, col
	// for stacktraces etc.
	function, err := value.NewFunction(def.Args, def.Body)
	if err != nil {
		return err
	}
	env.DefineGlobal(def.Name, function)
	return nil
}