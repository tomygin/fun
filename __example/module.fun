// ============================================================
// 多文件编程：包的导入 / 导出
//   import("路径")  读取并求值一个 .fun 文件，返回它导出的表
//   相对路径以"当前文件所在目录"为基准，扩展名 .fun 可省略
//   模块按绝对路径缓存，重复 import 只求值一次
//
// 设计哲学：模块就是一张表 { }。
//   隐式导出：模块顶层的所有名字自动成为导出表的字段。
//   显式导出：模块用 return { ... } 自己决定暴露什么。
// ============================================================

print("========== 1. 导入隐式导出的模块 ==========")

// mathlib 顶层定义的名字全部导出
math := import("pkg/mathlib.fun")
print("math.PI       =", math.PI)
print("math.add(2,3) =", math.add(2, 3))
print("math.max(7,4) =", math.max(7, 4))
print("math.abs(-9)  =", math.abs(-9))

print("========== 2. 导入显式导出的模块 ==========")

// stack 只导出 new，helper 是私有的
stack := import("pkg/stack")     // 省略 .fun 扩展名
print("stack 只导出:", keys(stack))
print("是否暴露 helper:", has(stack, "helper"))

s := stack.new()
s.push("a")
s.push("b")
s.push("c")
print("弹出:", s.pop())          // c
print("弹出:", s.pop())          // b
print("剩余:", s.size)           // 1

print("========== 3. 模块缓存 ==========")

// 重复 import 拿到的是同一张表
math2 := import("pkg/mathlib.fun")
print("同一模块缓存:", math2.PI == math.PI)
