package vm

import (
	"fmt"
	"maps"
	"reflect"
	"strconv"
)

// ----------------------------------------------------------------------------
// Environment
// ----------------------------------------------------------------------------

type Environment struct {
	local  map[string]any
	parent *Environment
}

func NewEnvironment(local map[string]any, parent *Environment) *Environment {
	if local == nil {
		local = make(map[string]any)
	}
	return &Environment{
		local:  local,
		parent: parent,
	}
}

func (env *Environment) String() string {
	return fmt.Sprintf("Environment: %v", env.local)
}

func (env *Environment) Lookup(name string) (any, error) {
	if value, exists := env.local[name]; exists {
		return value, nil
	}
	if env.parent != nil {
		return env.parent.Lookup(name)
	}
	return nil, fmt.Errorf(`"%s" is not found in Environment`, name)
}

func (env *Environment) Set(name string, value any) any {
	env.local[name] = value
	return value
}

func (env *Environment) Get(name string) any {
	return env.local[name]

}

func (env *Environment) Assign(name string, value any) (any, error) {
	if _, exists := env.local[name]; exists {
		return env.Set(name, value), nil
	}
	if env.parent != nil {
		return env.parent.Assign(name, value)
	}
	return nil, fmt.Errorf(`"%s" is not found in Environment`, name)
}

func (env *Environment) Clone() *Environment {
	newLocal := make(map[string]any)
	maps.Copy(newLocal, env.local)
	return NewEnvironment(newLocal, env.parent)
}

func (env *Environment) Next(local map[string]any) *Environment {
	cloned := env.Clone()
	return NewEnvironment(local, cloned)
}

func (env *Environment) Close() *Environment {
	if env.parent == nil {
		return env
	}
	return env.parent
}

// ----------------------------------------------------------------------------
// Tools
// ----------------------------------------------------------------------------

func isString(exp any) bool {
	if str, ok := exp.(string); ok {
		if len(str) >= 2 {
			return (str[0] == '"' && str[len(str)-1] == '"') ||
				(str[0] == '\'' && str[len(str)-1] == '\'')
		}
	}
	return false
}

func isNumber(exp any) bool {
	switch exp.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	}
	return false
}

func isVar(exp any) bool {
	if isNumber(exp) {
		return false
	}
	if _, ok := exp.([]any); ok {
		return false
	}
	if _, ok := exp.(string); ok {
		return !isString(exp)
	}
	return false
}

func isList(exp any) bool {
	_, ok := exp.([]any)
	return ok
}

func isDict(exp any) bool {
	_, ok := exp.(map[string]any)
	return ok
}

func toNumber(exp any) (float64, bool) {
	switch v := exp.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

func toBool(exp any) bool {
	switch v := exp.(type) {
	case bool:
		return v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Int() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	case string:
		return v != ""
	case []any:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	}
	return exp != nil
}

// ----------------------------------------------------------------------------
// VM
// ----------------------------------------------------------------------------

type VM struct {
	env *Environment
}

func NewVM() *VM {
	env := NewEnvironment(interfaceFunctions, nil)
	return &VM{env: env}
}

func (vm *VM) Call(exp any) (any, error) {
	return vm.Eval(exp)
}

func (vm *VM) Eval(exp any) (any, error) {
	// String literals
	if isString(exp) {
		str := exp.(string)
		return str[1 : len(str)-1], nil
	}

	// Numbers
	if isNumber(exp) {
		return exp, nil
	}

	// Variables
	if isVar(exp) {
		name := exp.(string)
		value, err := vm.env.Lookup(name)
		if err != nil {
			return nil, err
		}
		return value, nil
	}

	// Lists (expressions)
	if isList(exp) {
		expList := exp.([]any)

		// Empty statement
		if len(expList) == 0 {
			return nil, nil
		}

		operator := expList[0].(string)

		// Variable definition
		if operator == "var" {
			if len(expList) != 3 {
				return nil, fmt.Errorf("var requires exactly 2 arguments")
			}
			name := expList[1].(string)
			value, err := vm.Eval(expList[2])
			if err != nil {
				return nil, err
			}
			return vm.env.Set(name, value), nil
		}

		// Variable assignment
		if operator == "assign" {
			if len(expList) != 3 {
				return nil, fmt.Errorf("assign requires exactly 2 arguments")
			}
			name := expList[1].(string)
			value, err := vm.Eval(expList[2])
			if err != nil {
				return nil, err
			}
			result, err := vm.env.Assign(name, value)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		// 在 Eval 方法中，在 "Code block" 处理之前添加 return 处理：
		// Return statement
		if operator == "return" {
			// 符号表里面添加标志
			vm.env.Set("@ReturnFlag", true)

			if len(expList) == 1 {
				// return without value
				return nil, nil
			}

			resout, err := vm.Eval(expList[1])
			return resout, err

		}

		// Code block
		if operator == "begin" {
			vm.env = vm.env.Next(nil)
			var result any
			var err error

			for _, block := range expList[1:] {

				result, err = vm.Eval(block)
				if err != nil {
					vm.env = vm.env.Close()
					return nil, err
				}

				if isReturn, ok := vm.env.Get("@ReturnFlag").(bool); ok && isReturn {
					break
				}

			}
			vm.env = vm.env.Close()
			return result, nil
		}

		// If statement
		if operator == "if" {
			if len(expList) != 4 {
				return nil, fmt.Errorf("if requires exactly 3 arguments")
			}
			condition, err := vm.Eval(expList[1])
			if err != nil {
				return nil, err
			}

			if toBool(condition) {
				return vm.Eval(expList[2])
			} else {
				return vm.Eval(expList[3])
			}
		}

		// While loop
		if operator == "while" {
			if len(expList) != 3 {
				return nil, fmt.Errorf("while requires exactly 2 arguments")
			}
			var result any

			for {
				condition, err := vm.Eval(expList[1])
				if err != nil {
					return nil, err
				}
				if !toBool(condition) {
					break
				}
				result, err = vm.Eval(expList[2])
				if err != nil {
					return nil, err
				}
			}
			return result, nil
		}

		// Function definition
		if operator == "def" {
			if len(expList) != 4 {
				return nil, fmt.Errorf("def requires exactly 3 arguments")
			}
			name := expList[1].(string)
			params := expList[2].([]any)
			body := expList[3]

			fn := map[string]any{
				"params": params,
				"body":   body,
			}
			return vm.env.Set(name, fn), nil
		}

		// Function call
		fnValue, err := vm.env.Lookup(operator)
		if err != nil {
			return nil, err
		}

		// Evaluate arguments
		var args []any
		for _, arg := range expList[1:] {
			evalArg, err := vm.Eval(arg)
			if err != nil {
				return nil, err
			}
			args = append(args, evalArg)
		}

		// Built-in function
		if reflect.TypeOf(fnValue).Kind() == reflect.Func {
			fnReflect := reflect.ValueOf(fnValue)
			var reflectArgs []reflect.Value
			for _, arg := range args {
				reflectArgs = append(reflectArgs, reflect.ValueOf(arg))
			}
			results := fnReflect.Call(reflectArgs)
			if len(results) > 0 {
				return results[0].Interface(), nil
			}
			return nil, nil
		}

		// User-defined function
		if isDict(fnValue) {
			fnMap := fnValue.(map[string]any)
			params := fnMap["params"].([]any)
			body := fnMap["body"]

			// Create parameter bindings
			kv := make(map[string]any)
			for i, param := range params {
				if i < len(args) {
					kv[param.(string)] = args[i]
				}
			}

			vm.env = vm.env.Next(kv)
			result, err := vm.Eval(body)
			vm.env = vm.env.Close()
			return result, err
		}

		return nil, fmt.Errorf("unknown function: %s", operator)
	}

	return nil, fmt.Errorf("unknown expression: %v", exp)
}
