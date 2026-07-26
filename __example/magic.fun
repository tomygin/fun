// ============================================================
// 语言特色：管道 | 、链式调用、以及 { } 能玩出的花样
// ============================================================

print("========== 1. 管道 | ：数据从左流向右 ==========")

fun double(x) { return x * 2 }
fun plus(a, b) { return a + b }
fun square(x) { return x * x }

// x | f 就是 f(x)；x | g(a) 就是 g(x, a) —— 左值插为第一个实参
print(5 | double)                    // 10
print(5 | double | plus(3))          // 13
print(2 | plus(3) | square | str | len)   // "25" 的长度 = 2

// 管道让"先算什么后算什么"按阅读顺序排列
print("hello fun" | upper)           // HELLO FUN

// 管道也能流进对象的方法（this 照常绑定）
wrap := { exclaim: fun(s) { return s + "!!!" } }
print("nice" | wrap.exclaim)

print("========== 2. 链式调用：方法返回 this ==========")

fun newQuery() {
    return {
        parts: {},
        n: 0,
        where:   fun(c) { this.parts[this.n] = "WHERE " + c    this.n = this.n + 1  return this },
        orderBy: fun(c) { this.parts[this.n] = "ORDER BY " + c  this.n = this.n + 1  return this },
        limit:   fun(k) { this.parts[this.n] = "LIMIT " + str(k)  this.n = this.n + 1  return this },
        build:   fun() {
            sql := "SELECT *"
            for i := 0; i < this.n; i++ { sql = sql + " " + this.parts[i] }
            return sql
        }
    }
}

sql := newQuery().where("age > 18").orderBy("name").limit(10).build()
print(sql)

print("========== 3. 分发表：用 { } 替代 switch/case ==========")

// 其他语言要写一长串 switch；这里"分支"就是表里的一条数据，
// 可以运行时增删，还能被遍历。
handlers := {
    start: fun(who) { print("  ", who, "启动了服务") },
    stop:  fun(who) { print("  ", who, "停止了服务") },
    ping:  fun(who) { print("  ", who, ": pong") }
}

fun dispatch(cmd, who) {
    if has(handlers, cmd) {
        handlers[cmd](who)         // 从表里取出函数直接调用
    } else {
        print("  未知命令:", cmd)
    }
}

dispatch("start", "Gin")
dispatch("ping", "Tom")
dispatch("dance", "Bob")
// 运行时热插新命令 —— switch 做不到
handlers["dance"] = fun(who) { print("  ", who, "跳了一段舞") }
dispatch("dance", "Bob")

print("========== 4. 事件系统：几行代码的发布/订阅 ==========")

fun newEventBus() {
    return {
        subs: {},
        on: fun(event, handler) {
            if !has(this.subs, event) { this.subs[event] = { n: 0 } }
            list := this.subs[event]
            list[list.n] = handler
            list.n = list.n + 1
            return this
        },
        emit: fun(event, data) {
            if !has(this.subs, event) { return this }
            list := this.subs[event]
            for i := 0; i < list.n; i++ { list[i](data) }
            return this
        }
    }
}

bus := newEventBus()
bus.on("login", fun(user) { print("  欢迎,", user) })
   .on("login", fun(user) { print("  记录日志: ", user, " 登录了") })
   .on("logout", fun(user) { print("  再见,", user) })

bus.emit("login", "Gin").emit("logout", "Gin")

print("========== 5. 数组字面量与数据管道 ==========")

// { } 直接写数组（自动编号 0,1,2...），还能和键值对混用
nums := { 3, 1, 4, 1, 5, 9, 2, 6 }
print("原始:", nums)

// 自己写 map / filter —— 函数是值，表是数组，一切都是搭积木
fun mapT(t, f) {
    out := {}
    for i := 0; i < len(t); i++ { out[i] = f(t[i]) }
    return out
}
fun filterT(t, pred) {
    out := {}
    n := 0
    for i := 0; i < len(t); i++ {
        if pred(t[i]) { out[n] = t[i]  n = n + 1 }
    }
    return out
}
fun sumT(t) {
    s := 0
    for i := 0; i < len(t); i++ { s = s + t[i] }
    return s
}

// 管道 + 高阶函数 = 声明式数据流水线
result := nums
    | filterT(fun(x) { return x % 2 != 0 })    // 留奇数
    | mapT(fun(x) { return x * x })            // 平方
    | sumT                                     // 求和
print("奇数平方和:", result)                    // 9+1+1+25+81 = 117

print("========== 6. 表描述一切：配置即代码 ==========")

// 嵌套表 + 函数值：配置、校验规则、路由 …… 一张表全描述，
// 静态语言里要定义一堆类型才做得到。
app := {
    name: "fun-server",
    port: 8080,
    routes: {
        hello: fun(who) { return "hello, " + who },
        time:  fun(who) { return who + ", 现在是 " + now() }
    },
    check: fun() {
        if this.port > 0 && this.port < 65536 { return "配置合法" }
        return "端口非法"
    }
}
print(app.name, " -> ", app.check())
print(app.routes.hello("world"))

print("魔法演示结束 ✨")
