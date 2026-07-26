package vm

import "fmt"

// ----------------------------------------------------------------------------
// 协程（coroutine）—— 借鉴 lua 的对称式协程 API：
//   co := coroutine(fn)     用一个函数创建协程（此时并不运行）
//   resume(co, args...)     启动 / 恢复协程，返回协程 yield 出来的值
//   yield(v)                在协程内部让出，v 交给 resume；下次 resume 的
//                           第一个参数会成为 yield 的返回值
//   costatus(co)            "suspended" / "running" / "dead"
//
// 实现思路：树遍历解释器无法直接挂起 Go 调用栈，于是借助宿主的 goroutine。
// 每个协程是一个 goroutine，用两条无缓冲通道与主线程做严格交替的"接力"：
// 任一时刻只有一方在跑（协作式，非并行），因此可以安全地共享 vm.env ——
// 只需在每次交接处保存/恢复各自的执行环境即可。
// ----------------------------------------------------------------------------

// yieldMsg 是协程交回控制权时带出的消息
type yieldMsg struct {
	value any
	err   error
	done  bool // 协程整体结束（函数 return）时为 true
}

// Coroutine 一个协程
type Coroutine struct {
	fn       map[string]any // 协程体（用户函数）
	resumeCh chan []any     // 主 -> 协程：resume 传入的参数
	yieldCh  chan yieldMsg  // 协程 -> 主：yield / 结束时带出的值
	status   string         // "suspended" / "running" / "dead"
	started  bool
}

// coroutineCreate 实现 coroutine(fn)
func (vm *VM) coroutineCreate(args []any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("coroutine requires a function argument")
	}
	fn, ok := args[0].(map[string]any)
	if !ok || !isUserFunc(fn) {
		return nil, fmt.Errorf("coroutine argument must be a function")
	}
	return &Coroutine{
		fn:       fn,
		resumeCh: make(chan []any),
		yieldCh:  make(chan yieldMsg),
		status:   "suspended",
	}, nil
}

// coroutineResume 实现 resume(co, args...)
func (vm *VM) coroutineResume(args []any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("resume requires a coroutine argument")
	}
	co, ok := args[0].(*Coroutine)
	if !ok {
		return nil, fmt.Errorf("resume argument must be a coroutine")
	}
	if co.status == "dead" {
		return nil, fmt.Errorf("cannot resume a dead coroutine")
	}
	if co.status == "running" {
		return nil, fmt.Errorf("cannot resume a running coroutine")
	}

	resumeArgs := args[1:]

	// 保存主（或外层协程）的执行上下文
	savedEnv, savedReturning, savedCurrent := vm.env, vm.returning, vm.current

	// 首次 resume 时启动 goroutine
	if !co.started {
		co.started = true
		go func() {
			startArgs := <-co.resumeCh // 等待第一次 resume 的参数
			result, err := vm.runFunctionBody(co.fn, startArgs)
			co.status = "dead"
			co.yieldCh <- yieldMsg{value: result, err: err, done: true}
		}()
	}

	// 交接控制权给协程
	vm.current = co
	co.status = "running"
	co.resumeCh <- resumeArgs
	msg := <-co.yieldCh // 阻塞，直到协程 yield 或结束

	// 协程已把控制权交还，恢复调用方上下文
	vm.env, vm.returning, vm.current = savedEnv, savedReturning, savedCurrent

	if msg.err != nil {
		return nil, msg.err
	}
	if !msg.done {
		co.status = "suspended"
	}
	return msg.value, nil
}

// coroutineYield 实现 yield(v)
func (vm *VM) coroutineYield(args []any) (any, error) {
	co := vm.current
	if co == nil {
		return nil, fmt.Errorf("yield called outside of a coroutine")
	}

	var value any
	if len(args) > 0 {
		value = args[0]
	}

	// 保存协程自身的执行环境
	coEnv := vm.env
	co.status = "suspended"
	co.yieldCh <- yieldMsg{value: value, done: false}

	// 阻塞，直到再次被 resume
	resumeArgs := <-co.resumeCh
	// 重新拿到控制权，恢复协程环境
	vm.env = coEnv
	co.status = "running"

	if len(resumeArgs) > 0 {
		return resumeArgs[0], nil
	}
	return nil, nil
}

// runFunctionBody 在新环境里执行一个用户函数体（供协程使用）
func (vm *VM) runFunctionBody(fn map[string]any, args []any) (any, error) {
	params, _ := fn["params"].([]any)
	body := fn["body"]

	kv := make(map[string]any)
	for i, param := range params {
		if i < len(args) {
			kv[param.(string)] = args[i]
		}
	}

	parent := vm.global
	if ce, ok := fn["@ClosureEnv"].(*Environment); ok {
		parent = ce
	}

	vm.env = NewEnvironment(kv, parent)
	result, err := vm.Eval(body)
	vm.returning = false
	return result, err
}
