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
	if str, ok := exp.(string); ok {
		return !isString(str) && str != "this"
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

// anyType 是 interface{} 的反射类型，用于给内置函数传递 nil 参数
var anyType = reflect.TypeOf((*any)(nil)).Elem()

// buildReflectArgs 把求值后的参数转成反射值，nil 用 interface{} 的零值代替，
// 避免 reflect.ValueOf(nil) 产生非法 Value 而在调用时 panic。
func buildReflectArgs(args []any) []reflect.Value {
	reflectArgs := make([]reflect.Value, 0, len(args))
	for _, arg := range args {
		if arg == nil {
			reflectArgs = append(reflectArgs, reflect.Zero(anyType))
		} else {
			reflectArgs = append(reflectArgs, reflect.ValueOf(arg))
		}
	}
	return reflectArgs
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
	// returning 是函数返回信号：一旦 return 被求值就置为 true，
	// begin/while 会立即停止执行并向上冒泡，直到函数调用处消费它。
	returning bool
}

func NewVM() *VM {
	env := NewEnvironment(interfaceFunctions, nil)
	return &VM{env: env}
}

func (vm *VM) Call(exp any) (any, error) {
	return vm.Eval(exp)
}

func (vm *VM) Eval(exp any) (any, error) {
	// nil literal（没有值）
	if exp == nil {
		return nil, nil
	}

	// Boolean literals
	if b, ok := exp.(bool); ok {
		return b, nil
	}

	// String literals
	if isString(exp) {
		str := exp.(string)
		return str[1 : len(str)-1], nil
	}

	// Numbers
	if isNumber(exp) {
		return exp, nil
	}

	// this keyword - 返回当前环境的本地变量（跳过 @ 开头的内部变量）
	if exp == "this" {
		thisObj := make(map[string]any)
		for k, v := range vm.env.local {
			if len(k) > 0 && k[0] == '@' {
				continue
			}
			thisObj[k] = v
		}
		return thisObj, nil
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

		// nil literal 节点
		if operator == "nil" {
			return nil, nil
		}

		// Logical NOT
		if operator == "not" {
			value, err := vm.Eval(expList[1])
			if err != nil {
				return nil, err
			}
			return !toBool(value), nil
		}

		// Logical AND（短路，返回决定结果的操作数）
		if operator == "and" {
			left, err := vm.Eval(expList[1])
			if err != nil {
				return nil, err
			}
			if !toBool(left) {
				return left, nil
			}
			return vm.Eval(expList[2])
		}

		// Logical OR（短路，返回决定结果的操作数）
		if operator == "or" {
			left, err := vm.Eval(expList[1])
			if err != nil {
				return nil, err
			}
			if toBool(left) {
				return left, nil
			}
			return vm.Eval(expList[2])
		}

		// Table literal { key: value ... }
		if operator == "table" {
			table := make(map[string]any)
			for i := 1; i+1 < len(expList); i += 2 {
				key, err := vm.Eval(expList[i])
				if err != nil {
					return nil, err
				}
				value, err := vm.Eval(expList[i+1])
				if err != nil {
					return nil, err
				}
				table[fmt.Sprint(key)] = value
			}
			return table, nil
		}

		// Index access obj[key]（表或字符串）
		if operator == "index" {
			obj, err := vm.Eval(expList[1])
			if err != nil {
				return nil, err
			}
			key, err := vm.Eval(expList[2])
			if err != nil {
				return nil, err
			}

			switch container := obj.(type) {
			case map[string]any:
				// 缺失的键返回 nil（借鉴 lua）
				return container[fmt.Sprint(key)], nil
			case string:
				if idx, ok := toNumber(key); ok {
					runes := []rune(container)
					i := int(idx)
					if i >= 0 && i < len(runes) {
						return string(runes[i]), nil
					}
					return nil, fmt.Errorf("string index out of range: %d", i)
				}
				return nil, fmt.Errorf("string index must be a number")
			}
			return nil, fmt.Errorf("cannot index value of type %T", obj)
		}

		// Property access
		if operator == "property" {
			if len(expList) != 3 {
				return nil, fmt.Errorf("property requires exactly 2 arguments")
			}

			obj, err := vm.Eval(expList[1])
			if err != nil {
				return nil, err
			}

			propertyName := expList[2].(string)

			if objMap, ok := obj.(map[string]any); ok {
				if value, exists := objMap[propertyName]; exists {
					return value, nil
				}
				return nil, fmt.Errorf("property '%s' not found", propertyName)
			}

			return nil, fmt.Errorf("cannot access property on non-object")
		}

		// Method call
		if operator == "method-call" {
			if len(expList) < 3 {
				return nil, fmt.Errorf("method-call requires at least 2 arguments")
			}

			obj, err := vm.Eval(expList[1])
			if err != nil {
				return nil, err
			}

			methodName := expList[2].(string)

			// Evaluate arguments
			var args []any
			for _, arg := range expList[3:] {
				evalArg, err := vm.Eval(arg)
				if err != nil {
					return nil, err
				}
				args = append(args, evalArg)
			}

			if objMap, ok := obj.(map[string]any); ok {
				if method, exists := objMap[methodName]; exists {
					// 如果是用户定义的函数
					if methodMap, ok := method.(map[string]any); ok {
						params := methodMap["params"].([]any)
						body := methodMap["body"]

						// 创建新的环境，绑定参数
						kv := make(map[string]any)
						for i, param := range params {
							if i < len(args) {
								kv[param.(string)] = args[i]
							}
						}

						// 如果函数有闭包环境，则使用闭包环境作为父环境
						var parentEnv *Environment
						if closureEnv, exists := methodMap["@ClosureEnv"]; exists {
							parentEnv = closureEnv.(*Environment)
						} else {
							parentEnv = vm.env
						}

						// 创建新环境并执行
						oldEnv := vm.env
						vm.env = NewEnvironment(kv, parentEnv)

						result, err := vm.Eval(body)
						vm.env = oldEnv
						vm.returning = false // 消费 return 信号，函数边界到此为止
						return result, err
					}

					// 如果是内置函数
					if reflect.TypeOf(method).Kind() == reflect.Func {
						fnReflect := reflect.ValueOf(method)
						results := fnReflect.Call(buildReflectArgs(args))
						if len(results) > 0 {
							return results[0].Interface(), nil
						}
						return nil, nil
					}
				}
				return nil, fmt.Errorf("method '%s' not found", methodName)
			}

			return nil, fmt.Errorf("cannot call method on non-object")
		}

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

		// Variable assignment（支持 a = v / a.b = v / a[i] = v）
		if operator == "assign" {
			if len(expList) != 3 {
				return nil, fmt.Errorf("assign requires exactly 2 arguments")
			}

			value, err := vm.Eval(expList[2])
			if err != nil {
				return nil, err
			}

			// 简单变量赋值
			if name, ok := expList[1].(string); ok {
				return vm.env.Assign(name, value)
			}

			// 属性/下标赋值
			target, ok := expList[1].([]any)
			if !ok || len(target) != 3 {
				return nil, fmt.Errorf("invalid assignment target")
			}

			obj, err := vm.Eval(target[1])
			if err != nil {
				return nil, err
			}
			table, ok := obj.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("cannot assign to a non-table value")
			}

			var key string
			if target[0] == "property" {
				key = target[2].(string)
			} else { // index
				k, err := vm.Eval(target[2])
				if err != nil {
					return nil, err
				}
				key = fmt.Sprint(k)
			}
			table[key] = value
			return value, nil
		}

		// Return statement
		if operator == "return" {
			// 注意：先求值返回表达式再置位信号。
			// 因为求值过程可能包含函数调用，而函数调用返回时会清除信号，
			// 若提前置位会被内层调用清掉。
			var result any
			var err error
			if len(expList) >= 2 {
				result, err = vm.Eval(expList[1])
				if err != nil {
					return nil, err
				}
			}
			vm.returning = true
			return result, nil
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

				// 遇到 return 立即停止当前块，向上冒泡
				if vm.returning {
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
				// 循环体里的 return 需要跳出循环并继续向上冒泡
				if vm.returning {
					break
				}
			}
			return result, nil
		}

		// Function definition
		if operator == "fun" {
			if len(expList) != 4 {
				return nil, fmt.Errorf("fun requires exactly 3 arguments")
			}
			name := expList[1].(string)
			params := expList[2].([]any)
			body := expList[3]

			fn := map[string]any{
				"params": params,
				"body":   body,
				// 捕获定义时的环境作为闭包父环境。
				// 引用当前环境（而非快照）让函数能看到自己，从而支持递归，
				// 也让同一作用域内先后定义的函数彼此可见（相互递归）。
				"@ClosureEnv": vm.env,
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
			results := fnReflect.Call(buildReflectArgs(args))
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

			// 如果函数有闭包环境，则使用闭包环境作为父环境
			var parentEnv *Environment
			if closureEnv, exists := fnMap["@ClosureEnv"]; exists {
				parentEnv = closureEnv.(*Environment)
			} else {
				parentEnv = vm.env
			}

			// 创建新环境并执行函数
			oldEnv := vm.env
			vm.env = NewEnvironment(kv, parentEnv)
			result, err := vm.Eval(body)
			vm.env = oldEnv
			vm.returning = false // 消费 return 信号，函数边界到此为止

			return result, err
		}

		return nil, fmt.Errorf("unknown function: %s", operator)
	}

	return nil, fmt.Errorf("unknown expression: %v", exp)
}
