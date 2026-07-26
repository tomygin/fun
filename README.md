# Fun

一门比 Lua 还小的编程语言 —— 源码加测试大约 1,000 行，语法借鉴 Go 与 Python，
唯一的数据结构 `{ }` 借鉴 Lua。

## 介绍

很高兴为大家介绍这门小型编程语言，事实上这并不是我第一次尝试自己纯手动制作编程语
言，龙书和虎书这类专业书籍过于复杂，导致我经常半途而废，后来在朋友的推荐下接触了
lisp 实现来一个简易的解释器后端，然后笔者发现 lisp 方言解释器和其他的编程语言解
释器后端很像 ，lisp 的 s-expression 不就是手写语法树了嘛。然后正式迈入了 Fun 的
制作 …

Fun 这是一门非常小的编程语言，为什么我不说他是解释语言呢？如果你看过源码就知道
目前是 v1 版本，谁知道 v2 版本会不会有 go 或者 rust 实现的指令集或者编译器呢。

Fun 的速度不会比宿主语言更快这是上限，但是会更灵活。

Fun 是语言无关的编程语言,源码极其短小,你可以使用任何一门你熟悉的编程语言,用几天
的时间实现你自己的编程语言。

Fun 的语法来源于 go 和 python 当然唯一的数据结构 `{ }` 则是借鉴了 lua ，同样也是多文件编程的基础。

Fun 的实现思路,前端使用正则表达式提取token,然后左递归实现构建语法树(s-expression),然后运行在 lisp 的方言解释器上。

还在递归更新中 ...

总之，这是一门比 lua 还小的语言，源码加上测试代码目前在 1,000 行左右，希望你玩得开心。

## 语言特色

- **极小**：词法 + 语法 + 解释器总共约 1,000 行 Go 代码，一眼能读完。
- **语言无关**：前端正则取词 → 左递归构建 s-expression → Lisp 方言解释器，
  任何一门你熟悉的语言都能照着重写一遍。
- **语法混血**：`:=` 声明来自 Go，`for` 是唯一的循环关键字（同样来自 Go），
  真值/短路逻辑的味道来自 Python。
- **一种数据结构，无穷用法**：`{ }` 表借鉴自 Lua，既是记录（结构体）、
  数组，也是函数容器和对象 —— 一种结构撑起全部。
- **函数是一等公民**：支持匿名函数、闭包、递归、相互递归、柯里化调用
  `f(a)(b)`；把函数放进表就得到方法，方法里的 `this` 指向对象自身。
- **完整的面向对象**：不加一个关键字，`clone` + `merge` + `this` 就凑齐了
  类、实例、构造器、继承、重写、多态、super、mixin 和闭包私有成员。
- **管道 `|`**：`x | f | g(a)` 让数据从左流向右，纯语法糖，解析期改写成
  普通调用（借鉴 elixir）。
- **开放的元编程 `@` 空间**：运算符不是语法而是函数 —— `a + b` 就是
  `@add(a, b)`。`@` 空间对用户开放：运算符可以直接调用、当值传递，
  还能在局部作用域覆盖（作用域退出自动恢复）。
- **模板字符串**：`` `hello ${name}` `` 反引号字符串跨多行、原样保留、
  支持任意表达式插值；单双引号字符串支持 `\n` `\t` 等转义。
- **协程**：借鉴 lua 的 `coroutine` / `resume` / `yield`，可写生成器、
  做协作式调度（基于宿主 goroutine，协作式切换）。
- **多文件编程**：`import("x.fun")` 把一个文件当模块加载，模块就是一张
  导出表 —— 隐式导出全部顶层名字，或用 `return { ... }` 显式导出。

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
| 字符串 | `"hi"`、`'hi'`（支持转义）、`` `hi ${name}` ``（模板字符串） |
| 布尔 | `true`、`false` |
| 空 | `nil` |
| 表 | `{ name: "Gin", age: 18 }`、`{ 1, 2, 3 }` |

### 字符串：转义与模板

单/双引号字符串支持转义：`\n` `\t` `\r` `\\` `\"` `\'` `` \` `` `\0`
（未知转义原样保留）。

```js
print("第一行\n第二行")
print("他说: \"你好\"")
```

反引号是**模板字符串**：可跨多行、不处理转义（原样保留），
`${}` 里可以写**任意表达式** —— 变量、运算、函数与方法调用都行：

```js
name := "Gin"
print(`我是 ${name}，明年 ${age + 1} 岁`)
print(`调方法: ${user.greet()}`)
print(`多行
文本`)
print(`这里 \n 不转义，原样输出`)
```

### 真值规则：哪些会被判断为假

条件判断（`if` / `for` / `&&` / `||` / `!` / `bool()`）里，
**只有下面 5 类值是假（falsy）**：

| 假值 | 说明 |
| --- | --- |
| `false` | 布尔假 |
| `nil` | 空值（包括不存在的表键） |
| `0` / `0.0` | 数字零 |
| `""` | 空字符串 |
| `{}` | 空表 |

**除此之外一切都是真**，包括容易误会的：字符串 `"0"`、字符串 `"false"`、
负数 `-1`、非空表、所有函数。

```js
if "0" { print("字符串 '0' 是真！") }
port := 0
print(port || 8080)        // 8080 —— 因为 0 是假
```

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

// 管道：x | f 即 f(x)，优先级最低（详见"管道"一节）
print(5 | double | plus(3))
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

### 匿名函数（函数是值）

`fun(args){ ... }` 是一个表达式，可以赋值、当参数、当返回值，也可以放进表里：

```js
inc := fun(x) { return x + 1 }
print(inc(41))                 // 42

// 高阶函数
fun apply(f, x) { return f(x) }
print(apply(fun(n) { return n * n }, 9))   // 81
```

### 表：唯一的数据结构

`{ }` 是 Fun 里唯一的复合类型，但它非常灵活 —— 记录、数组、函数容器、对象，
全靠它。（完整演示见 [`__example/objects.fun`](__example/objects.fun)）

**1) 当作记录 / 结构体**

```js
user := { name: "Gin", age: 18, langs: { first: "go" } }
print(user.name)          // 属性访问
print(user["age"])        // 下标访问
print(user.langs.first)   // 支持嵌套
user.age = 19             // 属性赋值
user["city"] = "Hangzhou" // 下标赋值（动态新增字段）
```

**2) 当作数组（数组式字面量 + 数字下标 + `len` + `for` 遍历）**

```js
arr := { 10, 20, 30 }        // 数组式字面量，自动编号 0,1,2...
arr[3] = 40                  // 追加
for i := 0; i < len(arr); i++ {
    print(arr[i])
}
mixed := { "a", "b", name: "Gin" }   // 数组元素和键值对还能混写
```

**3) 把函数放进表**

```js
fun square(x) { return x * x }
t := {
    square: square,                          // 具名函数
    cube:   fun(x) { return x * x * x }       // 匿名函数
}
print(t.square(4), t.cube(3))                // 16 27
```

### 用 this 实现对象

表里的函数就是"方法"：用 `obj.method()` 调用时，方法内的 `this` 指向
`obj` 这张表本身，因此能读写自己的字段。这是 Fun 的对象模型。

```js
// 构造器：返回一张带方法的表
fun newAccount(owner, balance) {
    return {
        owner: owner,
        balance: balance,
        deposit: fun(n) {
            this.balance = this.balance + n   // this 指向被调用的这张表
            return this.balance
        },
        show: fun() { print(this.owner, "的余额:", this.balance) }
    }
}

acc := newAccount("Gin", 100)
acc.deposit(50)
acc.show()                 // Gin 的余额: 150
// 每次调用 newAccount 得到独立实例，状态互不影响
```

`this` 还有另一种玩法：在普通函数里（不是方法调用），`this` 会把当前作用域
打包成一张表。配合闭包，就能把状态藏在私有变量里：

```js
fun Counter(start) {
    count := start                            // 私有状态，外部访问不到
    fun inc() { count = count + 1  return count }
    fun value() { return count }
    return this                               // 打包 inc / value 成对象
}

c := Counter(10)
c.inc()
c.inc()
print(c.value())           // 12
```

### 完整的面向对象

不需要 `class` 关键字 —— `clone`（实例化）、`merge`（继承/mixin）加上
`this`，就凑齐了一门 OOP 语言的全部要素。（完整演示见
[`__example/oop.fun`](__example/oop.fun)）

```js
// 类 = 一张放着方法的原型表
Animal := {
    init:  fun(name) { this.name = name  return this },
    speak: fun() { print(this.name, ": ...") }
}

// 实例化 = clone；构造 = init 返回 this，可一行链完
cat := clone(Animal).init("小猫")

// 继承 = merge(clone(父类), { 覆盖与新增... })
// merge 优先级：后面的表覆盖前面的，所以子类字段一定赢过父类
Dog := merge(clone(Animal), {
    speak: fun() { print(this.name, ": 汪汪") },      // 重写
    fetch: fun() { print(this.name, "叼回飞盘")  return this }
})
dog := clone(Dog).init("旺财")

// 多态：同一条消息，各自的行为（鸭子类型）
animals := { cat, dog }
for i := 0; i < len(animals); i++ { animals[i].speak() }

// super：把父类方法挂到子类字段上，this 依旧晚绑定到实例
Puppy := merge(clone(Dog), {
    superSpeak: Animal.speak,
    speak: fun() { this.superSpeak()  print("（奶声奶气）") }
})

// mixin 多继承：能力就是表，merge 几张继承几种
CanSwim := { swim: fun() { print(this.name, "游泳")  return this } }
CanFly  := { fly:  fun() { print(this.name, "起飞")  return this } }
Duck := merge(clone(Animal), CanSwim, CanFly)

// 链式调用：方法返回 this 即可
clone(Duck).init("唐老鸭").swim().fly()

// 运行时改类：类只是表，随时打补丁
Animal.sleep = fun() { print(this.name, "睡了") }
```

### 管道 `|`

`x | f` 就是 `f(x)`；`x | g(a)` 就是 `g(x, a)` —— 左值插为第一个实参
（借鉴 elixir）。纯语法糖，解析期改写成普通调用。

```js
fun double(x) { return x * 2 }
fun plus(a, b) { return a + b }

print(5 | double | plus(3))       // 13
print("fun" | upper)              // FUN
print(3.7 | str | len)            // 3
print("wow" | obj.emphasize)      // 也能流进方法，this 照常绑定

// 配合高阶函数就是声明式数据流水线
result := nums
    | filterT(fun(x) { return x % 2 != 0 })
    | mapT(fun(x) { return x * x })
    | sumT
```

### `{ }` 的魔法

函数是值、表是唯一容器、值可直接调用（`handlers[cmd](x)`、`f(a)(b)`），
组合起来能做不少其他语言要动语法才能做的事。（完整演示见
[`__example/magic.fun`](__example/magic.fun)）

```js
// 分发表：用表替代 switch/case，分支可运行时热插拔
handlers := {
    start: fun(who) { print(who, "启动") },
    stop:  fun(who) { print(who, "停止") }
}
handlers["start"]("Gin")                       // 取出即调用
handlers["dance"] = fun(who) { print(who, "跳舞") }   // 热插新分支

// 几行代码的发布/订阅事件系统（见 magic.fun 第 4 节）
bus.on("login", fun(u) { print("欢迎", u) })
   .on("login", fun(u) { print("记日志", u) })
bus.emit("login", "Gin")

// 配置即代码：嵌套表 + 函数值描述一切
app := {
    port: 8080,
    routes: { hello: fun(w) { return "hello, " + w } },
    check: fun() { return this.port > 0 && this.port < 65536 }
}
```

### 元编程：开放的 `@` 空间

Fun 的运算符不是语法，是函数：`a + b` 在解析期被改写成 `@add(a, b)`，
然后像普通函数一样沿作用域链查找。`@` 空间对用户完全开放，由此得到
三种能力（完整演示见 [`__example/meta.fun`](__example/meta.fun)）：

```js
// 1) 运算符可以直接调用
print(@add(3, 4))              // 7

// 2) 运算符是一等值，可以当参数传
fun reduce(t, f, init) { /* ... */ }
reduce({ 1, 2, 3 }, @add, 0)   // 求和，不用写 fun(a,b){ return a+b }
reduce({ 1, 2, 3 }, @mul, 1)   // 求积

// 3) 局部覆盖操作符：在某个作用域内改写 + 的含义，退出自动恢复
fun vecDemo() {
    old := @add                          // 先保存原实现！
    @add := fun(a, b) {
        if type(a) == "table" {
            return { x: old(a.x, b.x), y: old(a.y, b.y) }
        }
        return old(a, b)
    }
    return { x: 1, y: 2 } + { x: 10, y: 20 }   // {x: 11, y: 22}
}
print(1 + 2)                   // 3 —— 外面的 + 不受影响
```

运算符与内部名对照：`+`→`@add`　`-`→`@sub`　`*`→`@mul`　`/`→`@div`
`%`→`@mod`　`>`→`@gt`　`<`→`@lt`　`>=`→`@gte`　`<=`→`@lte`
`==`→`@eq`　`!=`→`@neq`

两条注意事项：

- 覆盖体内的运算符**也走新实现**（作用域链就是这么工作的），直接用会
  无限递归 —— 想用原实现，先 `old := @add` 存下来（save-and-wrap）。
- `:=` 是局部覆盖（推荐，退出作用域自动恢复）；`=` 会沿作用域链改到
  定义处（可能全局生效），慎用。
- `@` 开头的名字被视为"内部空间"：`print` / `len` / `keys` / `this`
  都会自动跳过它们。

### 协程

借鉴 lua 的对称式协程：`coroutine` 创建、`resume` 恢复、`yield` 让出、
`costatus` 查状态。任一时刻只有一方在跑（协作式，非并行）。

```js
fun gen() {
    yield(1)
    yield(2)
    return 3          // 最后一次 resume 拿到返回值
}
co := coroutine(gen)
print(resume(co))     // 1
print(resume(co))     // 2
print(costatus(co))   // suspended
print(resume(co))     // 3
print(costatus(co))   // dead
```

`yield` 是双向的：它送出的值给 `resume` 的返回值，而下一次 `resume` 的
参数会成为 `yield` 的返回值，可用来做生成器、协作式调度等。（完整演示见
[`__example/coroutine.fun`](__example/coroutine.fun)）

### 多文件编程：包的导入 / 导出

`import("x.fun")` 读取并求值一个文件，返回它的**导出表**。相对路径以
当前文件所在目录为基准，扩展名 `.fun` 可省略，模块按绝对路径缓存。

**模块就是一张表**，导出有两种方式：

```js
// ---- mathlib.fun （隐式导出：顶层所有名字都导出）----
PI := 3.14159
fun add(a, b) { return a + b }
fun max(a, b) { if a > b { return a }  return b }
```

```js
// ---- stack.fun （显式导出：return 决定暴露什么）----
fun helper() { return "私有，不导出" }
fun new() { return { items: {}, size: 0 /* ... */ } }
return { new: new }              // 只导出 new
```

```js
// ---- main.fun ----
math := import("mathlib.fun")
print(math.add(2, 3))            // 5
print(math.PI)                   // 3.14159

stack := import("stack")          // 省略 .fun
s := stack.new()
print(has(stack, "helper"))      // false（没导出）
```

（完整演示见 [`__example/module.fun`](__example/module.fun) 与 `__example/pkg/`）

### 内置函数

| 函数 | 说明 |
| --- | --- |
| `print(...)` | 打印一行，自动格式化各类值 |
| `len(x)` | 字符串长度 / 表的键数量 |
| `type(x)` | 返回 `number` `string` `bool` `table` `function` `nil` |
| `str(x)` / `num(x)` / `bool(x)` | 类型转换 |
| `keys(t)` | 返回表所有键组成的新表（配合 `len` 遍历） |
| `has(t, k)` / `del(t, k)` | 键是否存在 / 删除键 |
| `clone(t)` | 浅拷贝一张表（OOP 的"实例化"） |
| `merge(dst, src...)` | 把后面各表的键依次覆盖进第一张表并返回它（OOP 的"继承 / mixin"）。**优先级：越靠后的表越高** —— `merge(a, b, c)` 中同名键 c 覆盖 b、b 覆盖 a，以最后出现的为准 |
| `upper(s)` / `lower(s)` | 字符串大小写转换 |
| `assert(cond, msg)` | 断言，失败则报错 |
| `input(prompt)` | 从标准输入读取一行 |
| `now()` | 当前时间 |
| `import(path)` | 加载一个 `.fun` 模块，返回其导出表 |
| `coroutine(fn)` / `resume(co, ...)` / `yield(v)` / `costatus(co)` | 协程原语 |
| `VERSION` | 版本号常量 |

完整可运行的示例见 [`__example/showcase.fun`](__example/showcase.fun)。

## 实现原理

```
源码 ──正则词法──▶ tokens ──左递归──▶ s-expression(语法树) ──▶ lisp 方言解释器
      lexer/               parser/                              vm/
```

- `lexer/`：用一组正则按优先级从左到右吃 token。
- `parser/`：左递归下降，把 token 变成 `[]any` 形式的 s-expression，
  例如 `1 + 2 * 3` → `["@add", 1, ["@mul", 2, 3]]`（运算符编译成 `@` 空间
  里的函数调用 —— 普通函数名不会撞到它们，而显式写 `@add` 就进入了
  元编程空间，见"元编程"一节）。
- `vm/`：一个树遍历（tree-walking）解释器。`vm.go` 负责求值与作用域（环境链），
  `interface.go` 提供内置函数，`module.go` 实现 import，`coroutine.go`
  基于 goroutine 实现协程。

关于速度：解释器做过三处针对性优化 —— 进入代码块不再整表拷贝环境
（零拷贝父链）、内置函数直接类型断言调用（绕开 reflect）、`@` 运算符
走快速分发路径。粗测递归斐波那契快了约 3 倍。上限依然是"不会比宿主
语言快"，但小语言也该跑得体面。

## 目录结构

```
.
├── main.go            入口：读文件 → 词法 → 语法 → 求值
├── lexer/             词法分析器
├── parser/            语法分析器（构建 s-expression）
├── vm/                解释器
│   ├── vm.go          求值与作用域
│   ├── interface.go   内置函数
│   ├── module.go      多文件编程：import / 导出
│   └── coroutine.go   协程：coroutine / resume / yield
└── __example/         示例代码
    ├── base.fun       基础语法
    ├── fibonaci.fun   递归：斐波那契
    ├── this.fun       用 this 构造对象
    ├── objects.fun    表 { } 的灵活性：记录 / 数组 / 方法 / 对象
    ├── oop.fun        完整面向对象：继承 / 多态 / super / mixin / 封装
    ├── magic.fun      管道 | / 分发表 / 事件系统 / 数据流水线
    ├── meta.fun       元编程 @ 空间 / 字符串转义 / 模板字符串 / 真值规则
    ├── coroutine.fun  协程：生成器 / 双向通信 / 协作式调度
    ├── module.fun     多文件编程：导入 pkg/ 下的模块
    ├── pkg/           被导入的模块（mathlib.fun / stack.fun）
    └── showcase.fun   特性总览（推荐从这里开始）
```
