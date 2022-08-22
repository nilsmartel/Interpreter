package interpreter

import (
	"errors"
	"fmt"
	"interpreter/ast"
	"interpreter/value"
)

func Call(env *Env, f *value.Function, args []value.Object) (value.Object, error) {
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

	switch expr.(type) {
	case *ast.ClassDefinition:
		def := expr.(*ast.ClassDefinition)
		err := defineClass(env, def)
		if err != nil {
			return nil, err
		}
		return nil, nil

	case *ast.FunctionDefinition:
		def := expr.(*ast.FunctionDefinition)
		err := defineFunc(env, def)
		if err != nil {
			return nil, err
		}
		return nil, nil
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
	env.DefineGlobal(def.Name, &function)
	return nil
}
