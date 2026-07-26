// ============================================================
// 完整的面向对象编程 —— 全部用 表 { } + this 搭出来，不加任何关键字
//
//   类       = 一张放着方法的表（原型）
//   实例化   = clone(原型)             浅拷贝出实例
//   构造器   = init 方法返回 this      支持链式 clone(X).init(...)
//   继承     = merge(clone(父类), {...})
//   重写     = 子类同名字段覆盖父类
//   多态     = 同一消息，不同对象不同响应（鸭子类型）
//   super    = 把父类方法挂到子类字段上，this 依旧晚绑定到实例
//   mixin    = merge 多张表，即"多继承"
//   封装     = 闭包私有变量，外部彻底摸不到
// ============================================================

print("========== 1. 类与实例 ==========")

// "类"就是一张原型表
Animal := {
    init: fun(name, sound) {
        this.name = name
        this.sound = sound
        return this               // 返回自身，支持链式
    },
    speak: fun() {
        print("  ", this.name, ": ", this.sound)
    },
    intro: fun() {
        print("  我是", this.name)
        return this
    }
}

// clone 实例化 + init 构造，一行完成
cat := clone(Animal).init("小猫", "喵")
cat.speak()
cat.intro().speak()               // 方法返回 this 即可链式调用

print("========== 2. 继承与重写 ==========")

// 继承：把父类拷贝一份，再覆盖/新增字段
Dog := merge(clone(Animal), {
    // 重写 speak
    speak: fun() {
        print("  ", this.name, ": 汪汪! 汪汪!")
    },
    // 新增方法
    fetch: fun() {
        print("  ", this.name, "叼回了飞盘")
        return this
    }
})

dog := clone(Dog).init("旺财", "汪")
dog.speak()                        // 用的是重写后的版本
dog.fetch().fetch()                // 新增方法，也能链式

print("========== 3. 多态（鸭子类型）==========")

// 同一段循环，对不同对象发同一条消息，行为各不相同
animals := { cat, dog, clone(Animal).init("未知生物", "...") }
for i := 0; i < len(animals); i++ {
    animals[i].speak()
}

print("========== 4. super：调用父类实现 ==========")

// 把父类方法挂到子类的一个字段上：
// 通过 this.xxx() 调用时，this 依旧晚绑定到实例
Puppy := merge(clone(Dog), {
    superSpeak: Animal.speak,      // 保存父类(祖先)实现
    speak: fun() {
        this.superSpeak()          // 先按父类方式叫，this 还是这只小狗
        print("  （幼犬奶声奶气）")
    }
})
puppy := clone(Puppy).init("小旺", "嗷呜")
puppy.speak()

print("========== 5. mixin：多继承组合 ==========")

// 能力就是一张表，merge 几张就"继承"几种能力
CanSwim := { swim: fun() { print("  ", this.name, "游得飞快")  return this } }
CanFly  := { fly:  fun() { print("  ", this.name, "起飞了")    return this } }

Duck := merge(clone(Animal), CanSwim, CanFly)
duck := clone(Duck).init("唐老鸭", "嘎嘎")
duck.speak()
duck.swim().fly()                  // mixin 进来的方法同样可以链式

print("========== 6. 封装：闭包私有成员 ==========")

// 想要真正的私有？把状态藏进闭包，只暴露方法。
// balance 在外部无论如何都拿不到，只能走 deposit / show。
fun newSafeAccount(owner) {
    balance := 0                   // 私有！
    return {
        owner: owner,
        deposit: fun(n) {
            if n > 0 { balance = balance + n }
            return this
        },
        show: fun() {
            print("  ", this.owner, "的余额:", balance)
            return this
        }
    }
}

acc := newSafeAccount("Gin")
acc.deposit(100).deposit(50).show()
print("  外部能看到 balance 吗:", has(acc, "balance"))

print("========== 7. 运行时改类：动态语言的自由 ==========")

// 类只是表，随时可以打补丁；之后 clone 出的实例立即拥有新方法
Animal.sleep = fun() { print("  ", this.name, "睡着了 zzz") }
sheep := clone(Animal).init("小羊", "咩")
sheep.sleep()

print("OOP 演示结束 🎉")
