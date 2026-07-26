# Fn

一门比 Lua 还小的编程语言 —— 源码加测试大约 1,000 行，语法借鉴 Go 与 Python，
唯一的数据结构 `{ }` 借鉴 Lua。

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

## 语言特色

- **极小**：词法 + 语法 + 解释器总共约 1,000 行 Go 代码，一眼能读完。
- **语言无关**：前端正则取词 → 左递归构建 s-expression → Lisp 方言解释器，
  任何一门你熟悉的语言都能照着重写一遍。
- **语法混血**：`:=` 声明来自 Go，`for` 是唯一的循环关键字（同样来自 Go），
  真值/短路逻辑的味道来自 Python。
- **一种数据结构**：`{ }` 表借鉴自 Lua，既当记录（结构体）又当数组，
  也是"对象"和多文件编程的基础。
- **函数是一等公民**：支持闭包、递归、相互递归；`return this` 就能把当前
  作用域打包成一个带方法的对象。

## 运行

```bash
go build -o fun .
./fun __example/showcase.fun
```

不带参数运行会得到一句 `have fun`。跑一遍测试：

```bash
go test ./...
```

## 使用文档

### 变量

```js
a := 10        // := 声明
a = a + 1      // = 赋值（变量必须先声明）
```

### 数据类型

| 类型 | 例子 |
| --- | --- |
| 数字 | `1`、`-3`、`3.14`（内部统一按数值比较，`1 == 1.0` 为真） |
| 字符串 | `"hi"`、`'hi'`（单双引号等价，暂不支持转义） |
| 布尔 | `true`、`false` |
| 空 | `nil` |
| 表 | `{ name: "Gin", age: 18 }` |

### 运算符

```js
// 算术：+ - * / %，以及一元负号
print(1 + 2 * 3)      // 7
print(10 % 3)         // 1
print(-a)             // 一元负号

// 字符串用 + 拼接（任一侧是字符串即拼接）
print("hello, " + "fn")

// 比较：== != > < >= <=
// 逻辑：&& || !（&& 和 || 会短路，并返回决定结果的操作数）
ok := true && !false
port := 0 || 8080     // 返回 8080
```

### 分支

```js
fun grade(s) {
    if s >= 90 {
        return "A"
    } else if s >= 60 {   // 支持 else if 链
        return "B"
    } else {
        return "C"
    }
}
```

### 循环

`for` 是唯一的循环关键字，仿照 Go 提供三种写法：

```js
// 1) 三段式
for i := 0; i < 10; i++ {
    print(i)
}

// 2) 条件式（相当于其他语言的 while）
n := 1
for n < 1000 {
    n = n * 2
}

// 3) 无限循环（配合函数里的 return 跳出）
fun countTo(k) {
    c := 0
    for {
        c++
        if c == k { return c }
    }
}
```

### 函数与递归

```js
// 支持自我递归
fun fib(n) {
    if n < 2 { return n }
    return fib(n - 1) + fib(n - 2)
}

// 也支持相互递归
fun isEven(n) {
    if n == 0 { return true }
    return isOdd(n - 1)
}
fun isOdd(n) {
    if n == 0 { return false }
    return isEven(n - 1)
}
```

### 闭包与对象

函数会捕获定义时的环境形成闭包；`return this` 把当前作用域打包成一张表，
这就是 Fn 里的"对象"：

```js
fun Counter(start) {
    count := start
    fun inc() { count = count + 1  return count }
    fun value() { return count }
    return this
}

c := Counter(10)
c.inc()
c.inc()
print(c.value())      // 12
```

### 表：唯一的数据结构

```js
// 当作记录 / 结构体
user := { name: "Gin", age: 18, langs: { first: "go" } }
print(user.name)          // 属性访问
print(user["age"])        // 下标访问
print(user.langs.first)   // 支持嵌套
user.age = 19             // 属性赋值
user["city"] = "Hangzhou" // 下标赋值（新增字段）

// 当作数组：用数字下标
arr := {}
for i := 0; i < 5; i++ {
    arr[i] = i * i
}
print(len(arr))           // 5
```

### 内置函数

| 函数 | 说明 |
| --- | --- |
| `print(...)` | 打印一行，自动格式化各类值 |
| `len(x)` | 字符串长度 / 表的键数量 |
| `type(x)` | 返回 `number` `string` `bool` `table` `function` `nil` |
| `str(x)` / `num(x)` / `bool(x)` | 类型转换 |
| `keys(t)` | 返回表所有键组成的新表（配合 `len` 遍历） |
| `has(t, k)` / `del(t, k)` | 键是否存在 / 删除键 |
| `upper(s)` / `lower(s)` | 字符串大小写转换 |
| `assert(cond, msg)` | 断言，失败则报错 |
| `input(prompt)` | 从标准输入读取一行 |
| `now()` | 当前时间 |
| `VERSION` | 版本号常量 |

完整可运行的示例见 [`__example/showcase.fun`](__example/showcase.fun)。

## 实现原理

```
源码 ──正则词法──▶ tokens ──左递归──▶ s-expression(语法树) ──▶ lisp 方言解释器
      lexer/               parser/                              vm/
```

- `lexer/`：用一组正则按优先级从左到右吃 token。
- `parser/`：左递归下降，把 token 变成 `[]any` 形式的 s-expression，
  例如 `1 + 2 * 3` → `["add", 1, ["mul", 2, 3]]`。
- `vm/`：一个树遍历（tree-walking）解释器，`vm.go` 负责求值与作用域（环境链），
  `interface.go` 提供内置函数。

## 目录结构

```
.
├── main.go            入口：读文件 → 词法 → 语法 → 求值
├── lexer/             词法分析器
├── parser/            语法分析器（构建 s-expression）
├── vm/                解释器与内置函数
└── __example/         示例代码
    ├── base.fun       基础语法
    ├── fibonaci.fun   递归：斐波那契
    ├── this.fun       用 this 构造对象
    └── showcase.fun   特性总览（推荐从这里开始）
```
