// 模块 mathlib —— 隐式导出：顶层定义的所有名字都会成为模块的导出表
//
// 注意：即便把函数命名为 add / sub，也不会影响 + / - 运算，
// 因为内置运算符放在用户取不到的命名空间里。

PI := 3.14159

fun add(a, b) { return a + b }
fun sub(a, b) { return a - b }
fun max(a, b) {
    if a > b { return a }
    return b
}
fun abs(x) {
    if x < 0 { return -x }
    return x
}
