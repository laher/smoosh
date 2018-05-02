package object

import (
	"fmt"
)

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func (e *Environment) Export() map[string]interface{} {
	if e == nil {
		return map[string]interface{}{}
	}
	x := e.outer.Export()
	for k, v := range e.store {
		x[k] = getValue(v, 3)
	}
	return x
}

func getValue(obj Object, depth int) interface{} {
	if depth < 1 {
		return "[N/A]"
	}
	switch obj.Type() {
	case ERROR_OBJ, INTEGER_OBJ, BOOLEAN_OBJ, STRING_OBJ:
		return obj.Inspect()
		//		case FUNCTION_OBJ:
		//		case BUILTIN_OBJ:
	case ARRAY_OBJ:
		o := obj.(*Array)
		items := []interface{}{}
		for _, v := range o.Elements {
			items = append(items, getValue(v, depth-1))
		}
		return items
	case HASH_OBJ:
		o := obj.(*Hash)
		ret := map[string]interface{}{}
		for _, v := range o.Pairs {
			ret[fmt.Sprintf("%v", getValue(v.Key, depth-1))] = getValue(v.Value, depth-1)
		}
		return ret
	}
	return nil
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
