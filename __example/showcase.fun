// ============================================================
// Fn 语言特性总览 (showcase)
// 涵盖：变量 / 运算 / 布尔 / 逻辑 / 分支 / 循环 / 函数 / 递归
//       闭包 / 对象 / 表(唯一数据结构) / 内置函数
// ============================================================

print("========== 1. 变量与运算 ==========")

a := 10
b := 3
print("a + b = ", a + b)
print("a - b = ", a - b)
print("a * b = ", a * b)
print("a / b = ", a / b)
print("a % b = ", a % b)      // 取模
print("-a    = ", -a)         // 一元负号

name := "Fn"
print("hello, " + name)       // 字符串拼接

print("========== 2. 布尔 / nil / 逻辑 ==========")

print("true && false = ", true && false)
print("true || false = ", true || false)
print("!true         = ", !true)
print("nil           = ", nil)

nothing := nil
if nothing == nil {
    print("nothing 确实是 nil")
}

// && / || 会短路并返回决定结果的操作数（借鉴 python）
port := 0
print("port || 8080 = ", port || 8080)

print("========== 3. 分支：if / else if / else ==========")

fun grade(score) {
    if score >= 90 {
        return "A"
    } else if score >= 80 {
        return "B"
    } else if score >= 60 {
        return "C"
    } else {
        return "D"
    }
}

print("95 -> ", grade(95))
print("83 -> ", grade(83))
print("61 -> ", grade(61))
print("40 -> ", grade(40))

print("========== 4. 循环：for ==========")

// 三段式 for（go 风格）
sum := 0
for i := 1; i <= 100; i++ {
    sum = sum + i
}
print("1 + 2 + ... + 100 = ", sum)

// 条件式 for（相当于其他语言的 while）
n := 1
for n < 1000 {
    n = n * 2
}
print("第一个 >= 1000 的 2 的幂 = ", n)

print("========== 5. 函数与递归 ==========")

fun factorial(n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
print("factorial(5) = ", factorial(5))

// 相互递归
fun isEven(n) {
    if n == 0 { return true }
    return isOdd(n - 1)
}
fun isOdd(n) {
    if n == 0 { return false }
    return isEven(n - 1)
}
print("isEven(10) = ", isEven(10))
print("isOdd(10)  = ", isOdd(10))

print("========== 6. 闭包与对象 (this) ==========")

// 函数捕获定义时的环境形成闭包；
// return this 会把当前作用域打包成一张表 —— 这就是 Fn 里的"对象"。
fun Counter(start) {
    count := start
    fun inc() {
        count = count + 1
        return count
    }
    fun value() {
        return count
    }
    return this
}

c := Counter(10)
c.inc()
c.inc()
c.inc()
print("counter value = ", c.value())

print("========== 7. 表：唯一的数据结构 ==========")

// 7.1 当作记录 / 结构体
user := {
    name: "Gin",
    age:  18,
    langs: { first: "go", second: "lua" }
}
print("user.name        = ", user.name)
print('user["age"]      = ', user["age"])
print("user.langs.first = ", user.langs.first)

// 就地修改与新增字段
user.age = 19
user["city"] = "Hangzhou"
print("修改后: ", user)

// 7.2 当作数组（用数字下标）
fun makeSquares(n) {
    arr := {}
    for i := 0; i < n; i++ {
        arr[i] = i * i
    }
    return arr
}
squares := makeSquares(6)
print("squares 长度 = ", len(squares))
for j := 0; j < len(squares); j++ {
    print("  squares[", j, "] = ", squares[j])
}

print("========== 8. 内置函数 ==========")

print("len('hello')      = ", len("hello"))
print("type(123)         = ", type(123))
print("type('s')         = ", type("s"))
print("type(user)        = ", type(user))
print("type(factorial)   = ", type(factorial))
print("str(3.14)         = ", str(3.14))
print("num('42') + 8     = ", num("42") + 8)
print("upper('fn')       = ", upper("fn"))
print("lower('FN')       = ", lower("FN"))
print("has(user,'name')  = ", has(user, "name"))
print("VERSION           = ", VERSION)

// keys + len 遍历一张表
print("遍历 user 的字段：")
ks := keys(user)
for k := 0; k < len(ks); k++ {
    field := ks[k]
    print("  ", field, " = ", user[field])
}

// assert 断言
assert(factorial(5) == 120, "factorial 计算错误")
print("所有断言通过，have fun! 🎉")
