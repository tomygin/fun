// ============================================================
// 元编程 @ 空间 / 字符串转义 / 模板字符串 / 真值规则
//
// Fun 的运算符不是语法，是函数：a + b 会被编译成 @add(a, b)。
// @ 空间是开放的 —— 你可以读它、传它、在局部作用域覆盖它。
// ============================================================

print("========== 1. 运算符即函数 ==========")

// 直接调用
print("@add(3,4) =", @add(3, 4))
print("@mul(3,4) =", @mul(3, 4))

// 直接当值传递 —— 其他语言要写 (a,b)=>a+b，这里运算符本身就是值
fun reduce(t, f, init) {
    acc := init
    for i := 0; i < len(t); i++ { acc = f(acc, t[i]) }
    return acc
}
nums := { 5, 2, 8 }
print("求和:", reduce(nums, @add, 0))
print("求积:", reduce(nums, @mul, 1))

print("========== 2. 局部覆盖操作符 ==========")

// 作用域内 @add := ... 即可改写 + 的含义，退出作用域自动恢复。
// 注意：覆盖体内的 + 也会走新实现（会无限递归），
// 想用原实现，先把它存下来（save-and-wrap 模式）。
fun vecDemo() {
    old := @add                        // 保存原来的 +
    @add := fun(a, b) {
        if type(a) == "table" {        // 表 + 表：按字段相加
            return { x: old(a.x, b.x), y: old(a.y, b.y) }
        }
        return old(a, b)               // 其余照旧
    }
    v := { x: 1, y: 2 } + { x: 10, y: 20 }
    print("向量相加:", v)
    print("数字不受影响:", 1 + 2)
}
vecDemo()
print("作用域外自动恢复:", 1 + 2)

// 块级覆盖同理
if true {
    @mul := fun(a, b) { return "被劫持了" }
    print("块内 2*3 =", 2 * 3)
}
print("块外 2*3 =", 2 * 3)

print("========== 3. 字符串转义 ==========")

// 单/双引号字符串支持：\n \t \r \\ \" \' \` \0
print("第一行\n第二行")
print("列1\t列2\t列3")
print("他说: \"你好\"")
print('单引号里的 \' 也行')
print("反斜杠自己: \\")

print("========== 4. 模板字符串 ` ` ==========")

// 反引号字符串：跨多行、不处理转义（原样保留）、支持 ${表达式} 插值。
// ${} 里可以写任意表达式 —— 变量、运算、函数调用、方法调用。
name := "Gin"
age := 18
print(`我是 ${name}，明年 ${age + 1} 岁`)

u := { name: "Tom", greet: fun() { return "hi, " + this.name } }
print(`调方法: ${u.greet()}`)
print(`嵌套花括号: ${len({ 1, 2, 3 })} 个元素`)

print(`转义在这里失效：\n 就是反斜杠加 n`)

print(`多行文本
  第二行
  第三行`)

print("========== 5. 真值规则：哪些是假 ==========")

// 只有这 5 类是假：false、nil、0（含 0.0）、空字符串 ""、空表 {}
print("false      ->", bool(false))
print("nil        ->", bool(nil))
print("0          ->", bool(0))
print("0.0        ->", bool(0.0))
print(`"" (空串)  ->`, bool(""))
print("{} (空表)  ->", bool({}))
// 其余一切都是真 —— 包括容易误会的这些：
print(`"0" (字符串) ->`, bool("0"))
print(`"false"      ->`, bool("false"))
print("负数 -1      ->", bool(-1))
print("非空表       ->", bool({ 1 }))

print("元编程演示结束 🔮")
