// ============================================================
// 协程 (coroutine) —— 借鉴 lua 的对称式协程
//   coroutine(fn)     用函数创建协程（不立即运行）
//   resume(co, ...)   启动 / 恢复，返回协程 yield 出来的值
//   yield(v)          在协程内让出，v 交给 resume；
//                     下次 resume 的参数会成为 yield 的返回值
//   costatus(co)      "suspended" / "running" / "dead"
// ============================================================

print("========== 1. 生成器：逐个产出值 ==========")

fun gen() {
    yield(1)
    yield(2)
    yield(3)
    return "done"
}
co := coroutine(gen)
print("类型:", type(co))
print(resume(co))          // 1
print(resume(co))          // 2
print(resume(co))          // 3
print("状态:", costatus(co))   // suspended
print(resume(co))          // done（函数返回值）
print("状态:", costatus(co))   // dead

print("========== 2. 双向通信 ==========")

// yield 既能把值送出，也能从下一次 resume 收到值
fun echo() {
    x := yield("ready")
    print("  协程收到:", x)
    y := yield("again")
    print("  协程收到:", y)
}
e := coroutine(echo)
print(resume(e))           // ready
resume(e, "hello")         // 协程收到: hello
resume(e, "world")         // 协程收到: world

print("========== 3. 斐波那契生成器 ==========")

fun fibGen(n) {
    a := 0
    b := 1
    for i := 0; i < n; i++ {
        yield(a)
        t := a + b
        a = b
        b = t
    }
}

f := coroutine(fibGen)
v := resume(f, 10)             // 第一次 resume 传入 n，并拿到第一个值
seq := str(v)
for costatus(f) != "dead" {
    v = resume(f)
    if costatus(f) != "dead" {
        seq = seq + " " + str(v)
    }
}
print("fib:", seq)             // 0 1 1 2 3 5 8 13 21 34

print("========== 4. 两个协程交替（协作式调度）==========")

fun worker(name, times) {
    for i := 0; i < times; i++ {
        print("  ", name, "第", i, "步")
        yield(nil)
    }
}
a := coroutine(worker)
b := coroutine(worker)
resume(a, "A", 3)
resume(b, "B", 3)
for costatus(a) != "dead" || costatus(b) != "dead" {
    if costatus(a) != "dead" { resume(a) }
    if costatus(b) != "dead" { resume(b) }
}
print("调度结束")
