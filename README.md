# Fn

## 介绍

很高兴为大家介绍这门小型编程语言，事实上这并不是我第一次尝试自己纯手动制作编程语
言，龙书和虎书这类专业书籍过于复杂，导致我经常半途而废，后来在朋友的推荐下接触了
lisp 实现来一个简易的解释器后端，然后笔者发现 lisp 方言解释器和其他的编程语言解
释器后端很像 ，lisp 的 s-expression 不就是手写语法树了嘛。然后正式迈入了 Fn 的
制作 …

Fn 这是一门非常小的编程语言，为什么我不说他是解释语言呢？如果你看过源码就知道
目前是 v1 版本，谁知道 v2 版本会不会有 go 或者 rust 实现的指令集或者编译器呢。

Fn 的速度不会比宿主语言更快这是上限，但是会更灵活。

Fn 是语言无关的编程语言,源码极其短小,你可以使用任何一门你熟悉的编程语言,用几天
的时间实现你自己的编程语言。

Fn 的语法来源于 go 和 python 当然唯一的数据结构 `{ }` 则是借鉴了 lua ，同样也是多文件编程的基础。

Fn 的实现思路,前端使用正则表达式提取token,然后左递归实现构建语法树(s-expression),然后运行在 lisp 的方言解释器上。

还在递归更新中 ...

总之，这是一门比 lua 还小的语言，源码加上测试代码目前在 1,000 行左右，希望你玩得开心。

## 运行

```bash
go build -o fun .
./fun __example/showcase.fun
```

不带参数运行会得到一句 `have fun`。

## 语法速览

```js
// 变量：:= 声明，= 赋值
a := 10
a = a + 1

// 算术：+ - * / %，以及一元负号 -a；字符串用 + 拼接
print("1 + 2 * 3 =", 1 + 2 * 3)
print("hello, " + "fn")

// 布尔 / nil / 逻辑运算（and / or 会短路）
ok := true and not false
port := 0 or 8080          // 返回 8080

// 分支：支持 else if 链
fun grade(s) {
    if s >= 90 { return "A" }
    else if s >= 60 { return "B" }
    else { return "C" }
}

// 循环
i := 0
while i < 5 { i++ }

// 函数：现在支持递归与相互递归
fun fib(n) {
    if n < 2 { return n }
    return fib(n - 1) + fib(n - 2)
}

// 闭包与对象：return this 把当前作用域打包成一张表
fun Counter() {
    count := 0
    fun inc() { count = count + 1  return count }
    return this
}
c := Counter()
c.inc()

// 表 { } —— 唯一的数据结构，既是记录也是数组
user := { name: "Gin", age: 18 }
user.age = 19              // 属性赋值
user["city"] = "Hangzhou" // 下标赋值
arr := {}
arr[0] = "a"              // 用数字下标当数组
```

## 内置函数

| 函数 | 说明 |
| --- | --- |
| `print(...)` | 打印一行，自动格式化各类值 |
| `len(x)` | 字符串长度 / 表的键数量 |
| `type(x)` | 返回 `number` `string` `bool` `table` `function` `nil` |
| `str(x)` / `num(x)` / `bool(x)` | 类型转换 |
| `keys(t)` | 返回表所有键组成的新表（配合 `len` 遍历） |
| `has(t, k)` / `del(t, k)` | 键是否存在 / 删除键 |
| `upper(s)` / `lower(s)` | 大小写转换 |
| `assert(cond, msg)` | 断言，失败则报错 |
| `input(prompt)` | 从标准输入读取一行 |
| `now()` | 当前时间 |
| `VERSION` | 版本号常量 |

完整可运行的示例见 [`__example/showcase.fun`](__example/showcase.fun)。
