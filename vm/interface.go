package vm

import (
	"fmt"
	"reflect"
	"time"
)

// ----------------------------------------------------------------------------
// Interface (built-in functions)
// ----------------------------------------------------------------------------

var interfaceFunctions = map[string]any{
	"VERSION": "1.0.0",
	"neq": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		return !reflect.DeepEqual(args[0], args[1])
	},
	"eq": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		return reflect.DeepEqual(args[0], args[1])
	},
	"gt": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		if num1, ok := toNumber(args[0]); ok {
			if num2, ok := toNumber(args[1]); ok {
				return num1 > num2
			}
		}
		return false
	},
	"gte": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		if num1, ok := toNumber(args[0]); ok {
			if num2, ok := toNumber(args[1]); ok {
				return num1 >= num2
			}
		}
		return false
	},
	"lte": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		if num1, ok := toNumber(args[0]); ok {
			if num2, ok := toNumber(args[1]); ok {
				return num1 <= num2
			}
		}
		return false
	},
	"lt": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		if num1, ok := toNumber(args[0]); ok {
			if num2, ok := toNumber(args[1]); ok {
				return num1 < num2
			}
		}
		return false
	},
	"add": func(args ...any) any {
		if len(args) < 2 {
			return 0.0
		}
		if num1, ok := toNumber(args[0]); ok {
			if num2, ok := toNumber(args[1]); ok {
				return num1 + num2
			}
		}
		return 0.0
	},
	"sub": func(args ...any) any {
		if len(args) < 2 {
			return 0.0
		}
		if num1, ok := toNumber(args[0]); ok {
			if num2, ok := toNumber(args[1]); ok {
				return num1 - num2
			}
		}
		return 0.0
	},
	"mul": func(args ...any) any {
		if len(args) < 2 {
			return 1.0
		}
		if num1, ok := toNumber(args[0]); ok {
			if num2, ok := toNumber(args[1]); ok {
				return num1 * num2
			}
		}
		return 1.0
	},
	"div": func(args ...any) any {
		if len(args) < 2 {
			return 0.0
		}
		if num1, ok := toNumber(args[0]); ok {
			if num2, ok := toNumber(args[1]); ok && num2 != 0 {
				return num1 / num2
			}
		}
		return 0.0
	},
	"print": func(args ...any) any {
		for _, arg := range args {
			fmt.Print(arg)
		}
		fmt.Println()
		return nil
	},
	"now": func(args ...any) any {
		return time.Now().String()
	},
}
