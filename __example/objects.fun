// ============================================================
// 表 { } 的灵活性 —— Fun 唯一的数据结构
//   1. 定义表（记录 / 结构体）
//   2. 把函数放进表（具名 & 匿名）
//   3. 表当数组用
//   4. 用 this 实现对象：方法读写自身字段
//   5. 用 this 实现对象：闭包封装私有状态
// ============================================================

print("========== 1. 表：定义与访问 ==========")

// 表就是键值对集合，可嵌套
point := {
    x: 3,
    y: 4,
    label: "A"
}
print("point.x =", point.x)          // 属性访问
print("point['y'] =", point["y"])    // 下标访问
point.x = 30                          // 属性赋值
point["z"] = 5                        // 下标赋值（动态新增字段）
print("修改后:", point)

print("========== 2. 把函数放进表 ==========")

// 2.1 放具名函数
fun dist2(p) { return p.x * p.x + p.y * p.y }
math := { dist2: dist2 }
print("math.dist2(point) =", math.dist2(point))

// 2.2 放匿名函数（fun 作为表达式），并用 this 拿到"这张表自己"
vec := {
    x: 3,
    y: 4,
    // this 指向被调用的这张表，方法即可读自身字段
    length2: fun() { return this.x * this.x + this.y * this.y },
    scale: fun(k) {
        this.x = this.x * k
        this.y = this.y * k
        return this            // 返回自身，可链式调用
    }
}
print("vec.length2() =", vec.length2())
vec.scale(2)
print("vec.scale(2) 后:", vec)

print("========== 3. 表当数组用 ==========")

// 用数字下标即可把表当数组，len 得到元素个数
nums := {}
for i := 0; i < 5; i++ {
    nums[i] = i * 10
}
print("nums 长度 =", len(nums))
sum := 0
for i := 0; i < len(nums); i++ {
    sum = sum + nums[i]
}
print("nums 求和 =", sum)

print("========== 4. this 对象：方法读写字段 ==========")

// 构造器：返回一张带方法的表，方法用 this 操作字段
fun newAccount(owner, balance) {
    return {
        owner: owner,
        balance: balance,
        deposit: fun(n) {
            this.balance = this.balance + n
            return this.balance
        },
        withdraw: fun(n) {
            if n > this.balance {
                print("  ", this.owner, "余额不足")
                return this.balance
            }
            this.balance = this.balance - n
            return this.balance
        },
        show: fun() { print("  ", this.owner, "的余额:", this.balance) }
    }
}

gin := newAccount("Gin", 100)
tom := newAccount("Tom", 0)
gin.deposit(50)
gin.withdraw(30)
tom.deposit(999)
tom.withdraw(2000)   // 触发余额不足
// 两个实例状态互相独立
gin.show()
tom.show()

print("========== 5. this 对象：闭包封装私有状态 ==========")

// 另一种风格：状态藏在闭包变量里（外部访问不到 count），
// return this 把内部函数打包成对象的方法。
fun Counter(start) {
    count := start
    fun inc() { count = count + 1  return count }
    fun dec() { count = count - 1  return count }
    fun value() { return count }
    return this
}
c := Counter(10)
c.inc()
c.inc()
c.dec()
print("counter =", c.value())

print("========== 6. 综合：用 表+数组+this 实现一个栈 ==========")

fun newStack() {
    return {
        items: {},          // 底层用一张表当数组
        size: 0,
        push: fun(x) {
            this.items[this.size] = x
            this.size = this.size + 1
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

s := newStack()
s.push("a")
s.push("b")
s.push("c")
print("栈是否为空:", s.isEmpty())
print("弹出:", s.pop())
print("弹出:", s.pop())
print("栈内剩余数量:", s.size)
print("弹出:", s.pop())
print("栈是否为空:", s.isEmpty())
