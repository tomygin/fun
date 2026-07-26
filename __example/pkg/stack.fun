// 模块 stack —— 显式导出：用 return 决定对外暴露什么
//
// 这里只导出一个 new 构造器；helper 是私有的，外部访问不到。

fun helper() { return "内部实现，不导出" }

// 构造一个栈对象（基于表 + this）
fun new() {
    return {
        items: {},
        size: 0,
        push: fun(x) {
            this.items[this.size] = x
            this.size = this.size + 1
            return this
        },
        pop: fun() {
            if this.size == 0 { return nil }
            this.size = this.size - 1
            v := this.items[this.size]
            del(this.items, this.size)
            return v
        },
        isEmpty: fun() { return this.size == 0 }
    }
}

// 只导出 new
return { new: new }
