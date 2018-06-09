package stdlib

import (
	"fmt"

	"github.com/laher/smoosh/object"
)

// RegisterBuiltin registers a 'builtin'
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

// GetFn returns function if defined
func GetFn(name string) (*object.Builtin, bool) {
	bi, ok := builtins[name]
	if !ok {
		return nil, ok
	}
	return bi, ok
}

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(scope object.Scope, args ...object.Object) (object.Operation, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			var ret object.Object
			switch arg := args[0].(type) {
			case *object.Array:
				ret = &object.Integer{Value: int64(len(arg.Elements))}
			case *object.String:
				ret = &object.Integer{Value: int64(len(arg.Value))}
			default:
				return nil, fmt.Errorf("argument to `len` not supported, got %s",
					args[0].Type())
			}
			return func() object.Object {
				return ret
			}, nil
		},
	},
	"first": &object.Builtin{
		Fn: func(scope object.Scope, args ...object.Object) (object.Operation, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return nil, fmt.Errorf("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}
			arr := args[0].(*object.Array)
			return func() object.Object {
				if len(arr.Elements) > 0 {
					return arr.Elements[0]
				}
				return Null
			}, nil
		},
	},
	"last": &object.Builtin{
		Fn: func(scope object.Scope, args ...object.Object) (object.Operation, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return nil, fmt.Errorf("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			return func() object.Object {
				if length > 0 {
					return arr.Elements[length-1]
				}
				return Null
			}, nil
		},
	},
	"rest": &object.Builtin{
		Fn: func(scope object.Scope, args ...object.Object) (object.Operation, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return nil, fmt.Errorf("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			return func() object.Object {
				if length > 0 {
					newElements := make([]object.Object, length-1, length-1)
					copy(newElements, arr.Elements[1:length])
					return &object.Array{Elements: newElements}
				}
				return Null
			}, nil
		},
	},
	"push": &object.Builtin{
		Fn: func(scope object.Scope, args ...object.Object) (object.Operation, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return nil, fmt.Errorf("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			return func() object.Object {
				newElements := make([]object.Object, length+1, length+1)
				copy(newElements, arr.Elements)
				newElements[length] = args[1]
				return &object.Array{Elements: newElements}
			}, nil
		},
	},
}
