// ============================================================
// 错误处理：throw / try —— 没有新语法，错误就是值
//
//   throw(v)           抛出任意值（字符串、表 …… 都行）
//   try(fn, args...)   调用 fn 并捕获错误，返回结果表：
//                        成功 -> { ok: true,  value: 返回值 }
//                        失败 -> { ok: false, error: 错误值 }
//
// 结果表与 http 响应 { ok, ... } 同构 —— 全语言一种判错姿势。
// ============================================================

print("========== 1. 抛出与捕获 ==========")

fun withdraw(balance, n) {
    if n > balance {
        throw("余额不足")               // 错误就是一个值
    }
    return balance - n
}

r := try(withdraw, 100, 30)             // 直接把函数和参数交给 try
print("取 30:", r.ok, "余额", r.value)

r = try(withdraw, 100, 500)
print("取 500:", r.ok, "原因:", r.error)

print("========== 2. 错误可以是表（带结构的错误）==========")

fun readConfig(name) {
    throw({ code: 404, file: name, msg: "配置不存在" })
}
r = try(readConfig, "app.conf")
if !r.ok {
    print("错误码:", r.error.code, "文件:", r.error.file, "信息:", r.error.msg)
}

print("========== 3. try + 匿名函数 = 无语法的 try 块 ==========")

r = try(fun() {
    step1 := "准备"
    throw("第二步炸了")
    return step1 + "完成"               // 不会执行
})
print(r.ok, "|", r.error)

print("========== 4. 运行时错误同样能捕获 ==========")

r = try(fun() { return not_exist_var })
print("未定义变量:", r.error)

r = try(fun() { nope := 1  return nope.field.deep })
print("非法访问:", r.error)

r = try(fun() { assert(1 > 2, "1 竟然不大于 2") })
print("断言失败:", r.error)

r = try(fun() { return 1 / 0 })
print("除以零:", r.error)

print("========== 5. 错误穿透调用栈，直到最近的 try ==========")

fun layer3() { throw("最深处") }
fun layer2() { layer3() }
fun layer1() { layer2() }
r = try(layer1)
print("穿了三层:", r.error)

// 嵌套 try：内层捕获了，外层就是成功
r = try(fun() {
    inner := try(fun() { throw("小错误") })
    return "内层已处理: " + inner.error
})
print(r.ok, "|", r.value)

print("========== 6. 和 json / http 的搭配 ==========")

// json.decode 遇到坏数据会抛错 —— 用 try 稳稳接住
r = try(json.decode, "{这不是json}")
print("坏 JSON:", r.ok, "|", r.error)

// 合法的 null 和"解析失败"能明确区分
r = try(json.decode, "null")
print("合法 null:", r.ok, "| value =", r.value)

// json.encode 第二个参数是缩进（美化输出）
print(json.encode({ name: "Gin", tags: { "go", "lua" } }, 2))

print("========== 7. throw 也是一等值 ==========")

// 运算符是函数，throw 也是函数 —— 可以放进表、当参数传
fun mapT(t, f) {
    out := {}
    for i := 0; i < len(t); i++ { out[i] = f(t[i]) }
    return out
}
fun checkPositive(x) {
    if x < 0 { throw("发现负数: " + str(x)) }
    return x
}
r = try(mapT, { 1, 2, -3, 4 }, checkPositive)
print("校验数组:", r.ok, "|", r.error)

print("错误处理演示结束 🛡️")
