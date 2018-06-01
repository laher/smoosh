package stdlib

import (
	"fmt"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func RegisterBuiltin(name string, def *object.Builtin) {
	if _, ok := builtins[name]; ok {
		panic("fn '" + name + "' already defined")
	}
	builtins[name] = def
}

// RegisterFn registers a 'builtin function'
func RegisterFn(name string, def object.BuiltinFunction) {
	RegisterBuiltin(name, &object.Builtin{
		Fn: def,
	})
}

//GetFun returns function if defined
func GetFn(name string) (*object.Builtin, bool) {
	bi, ok := builtins[name]
	if !ok {
		return nil, ok
	}
	return bi, ok
}

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			if len(args) != 1 {
				return object.NewError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return object.NewError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
	},
	"puts": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return Null
		},
	},
	"first": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			if len(args) != 1 {
				return object.NewError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return object.NewError("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return Null
		},
	},
	"last": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			if len(args) != 1 {
				return object.NewError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return object.NewError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return Null
		},
	},
	"rest": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			if len(args) != 1 {
				return object.NewError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return object.NewError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object.Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])
				return &object.Array{Elements: newElements}
			}
			return Null
		},
	},
	"push": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			if len(args) != 2 {
				return object.NewError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return object.NewError("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &object.Array{Elements: newElements}
		},
	},
}
