package vm

import "fmt"

// ----------------------------------------------------------------------------
// 错误处理：throw / try
//
// 设计（参考 lua 的 pcall/error 与 elixir 的结果元组，贴合 Fun 哲学）：
//   - 不新增任何语法：throw 和 try 都只是函数。
//   - 错误就是一个值：throw("余额不足")、throw({ code: 400, msg: "..." })。
//   - try(fn, args...) 调用 fn 并捕获错误，返回结果表：
//       成功 -> { ok: true,  value: 返回值 }
//       失败 -> { ok: false, error: 错误值 }
//     与 http 响应表 { ok, ... } 同构，判断方式一致。
//   - 运行时错误（未定义变量、调用不可调用的值等）同样被 try 捕获，
//     错误值是描述字符串。未被捕获的错误一路冒泡到顶层，干净地报错退出。
// ----------------------------------------------------------------------------

// FunError 承载 throw 抛出的 Fun 值，走 VM 既有的 error 冒泡通道
type FunError struct {
	Value any
}

func (e *FunError) Error() string {
	return formatValue(e.Value)
}

// throwFunc 是内置函数 throw：抛出任意值作为错误。
// 采用 (any, error) 签名的内置函数，由 applyFunction 识别。
func throwFunc(args ...any) (any, error) {
	var v any = "error"
	if len(args) > 0 {
		v = args[0]
	}
	return nil, &FunError{Value: v}
}

// tryFunc 构造绑定 VM 的 try：调用函数并捕获一切错误（含宿主 panic）
func (vm *VM) tryFunc() func(...any) any {
	return func(args ...any) (result any) {
		if len(args) < 1 {
			return map[string]any{"ok": false, "error": "try requires a function"}
		}

		// 捕获宿主层 panic（如模板字符串里的语法错误等极端情况）
		defer func() {
			if r := recover(); r != nil {
				result = map[string]any{"ok": false, "error": fmt.Sprint(r)}
			}
		}()

		value, err := vm.applyFunction(args[0], args[1:])
		// try 的边界也是函数边界：清掉逃逸的 return 信号
		vm.returning = false
		if err != nil {
			if fe, ok := err.(*FunError); ok {
				return map[string]any{"ok": false, "error": fe.Value}
			}
			return map[string]any{"ok": false, "error": err.Error()}
		}
		return map[string]any{"ok": true, "value": value}
	}
}
