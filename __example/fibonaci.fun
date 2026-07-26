// 斐波那契数列 —— 递归现在可以正常工作了
// 函数在自己的闭包里可见，所以能自我调用（也支持相互递归）

fun fib(n) {
    if n < 2 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

// 打印前 10 项
i := 0
while i < 10 {
    print("fib(", i, ") = ", fib(i))
    i++
}
