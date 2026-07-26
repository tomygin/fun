package vm_test

import (
	"fun/lexer"
	"fun/parser"
	"fun/vm"
	"os"
	"path/filepath"
	"testing"
)

// run 把源码跑一遍，返回最后一个表达式的值
func run(t *testing.T, code string) any {
	t.Helper()
	tokens := lexer.NewLexer().Tokenize(code)
	ast := parser.NewParser().Parse(tokens)
	result, err := vm.NewVM().Call(ast)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	return result
}

func TestArithmetic(t *testing.T) {
	if got := run(t, `1 + 2 * 3`); got != 7.0 {
		t.Errorf("1 + 2 * 3 = %v, want 7", got)
	}
	if got := run(t, `10 % 3`); got != 1.0 {
		t.Errorf("10 %% 3 = %v, want 1", got)
	}
	if got := run(t, `-5 + 2`); got != -3.0 {
		t.Errorf("-5 + 2 = %v, want -3", got)
	}
}

func TestStringConcat(t *testing.T) {
	if got := run(t, `"hello, " + "world"`); got != "hello, world" {
		t.Errorf(`concat = %v, want "hello, world"`, got)
	}
}

func TestComparisonMixedNumbers(t *testing.T) {
	// 算术结果是 float，字面量是 int，两者必须能相等比较
	if got := run(t, `x := 10 - 10  x == 0`); got != true {
		t.Errorf("0.0 == 0 => %v, want true", got)
	}
}

func TestLogical(t *testing.T) {
	if got := run(t, `true && false`); got != false {
		t.Errorf("true && false = %v", got)
	}
	if got := run(t, `false || 42`); got != 42 {
		t.Errorf("false || 42 = %v, want 42", got)
	}
	if got := run(t, `!nil`); got != true {
		t.Errorf("!nil = %v, want true", got)
	}
	if got := run(t, `!(1 > 2) && 2 < 3`); got != true {
		t.Errorf("!(1 > 2) && 2 < 3 = %v, want true", got)
	}
}

// TestReturnInNestedBlock 覆盖修复的 return 冒泡问题：
// if 里的 return 必须终止整个函数，而不是继续往下执行。
func TestReturnInNestedBlock(t *testing.T) {
	code := `
	fun classify(n) {
		if n < 0 {
			return "neg"
		}
		return "non-neg"
	}
	classify(-5)`
	if got := run(t, code); got != "neg" {
		t.Errorf("classify(-5) = %v, want neg", got)
	}
}

// TestRecursion 覆盖修复的递归问题：函数在闭包里能看到自己。
func TestRecursion(t *testing.T) {
	code := `
	fun fib(n) {
		if n < 2 { return n }
		return fib(n - 1) + fib(n - 2)
	}
	fib(10)`
	if got := run(t, code); got != 55.0 {
		t.Errorf("fib(10) = %v, want 55", got)
	}
}

func TestMutualRecursion(t *testing.T) {
	code := `
	fun isEven(n) {
		if n == 0 { return true }
		return isOdd(n - 1)
	}
	fun isOdd(n) {
		if n == 0 { return false }
		return isEven(n - 1)
	}
	isEven(10)`
	if got := run(t, code); got != true {
		t.Errorf("isEven(10) = %v, want true", got)
	}
}

func TestTableLiteralAndAccess(t *testing.T) {
	if got := run(t, `u := { name: "Gin", age: 18 }  u.name`); got != "Gin" {
		t.Errorf("u.name = %v, want Gin", got)
	}
	if got := run(t, `u := { name: "Gin", age: 18 }  u["age"]`); got != 18 {
		t.Errorf(`u["age"] = %v, want 18`, got)
	}
}

func TestTableMutation(t *testing.T) {
	code := `
	u := { a: 1 }
	u.b = 2
	u["c"] = 3
	u.a + u.b + u.c`
	if got := run(t, code); got != 6.0 {
		t.Errorf("sum = %v, want 6", got)
	}
}

func TestTableAsArray(t *testing.T) {
	code := `
	arr := {}
	for i := 0; i < 5; i++ {
		arr[i] = i * i
	}
	arr[4]`
	if got := run(t, code); got != 16.0 {
		t.Errorf("arr[4] = %v, want 16", got)
	}
}

func TestForConditionForm(t *testing.T) {
	code := `
	n := 1
	for n < 100 {
		n = n * 2
	}
	n`
	if got := run(t, code); got != 128.0 {
		t.Errorf("for-cond result = %v, want 128", got)
	}
}

func TestForThreeClause(t *testing.T) {
	code := `
	sum := 0
	for i := 1; i <= 100; i++ {
		sum = sum + i
	}
	sum`
	if got := run(t, code); got != 5050.0 {
		t.Errorf("sum 1..100 = %v, want 5050", got)
	}
}

func TestForInfiniteWithReturn(t *testing.T) {
	code := `
	fun countTo(k) {
		c := 0
		for {
			c++
			if c == k { return c }
		}
	}
	countTo(7)`
	if got := run(t, code); got != 7.0 {
		t.Errorf("countTo(7) = %v, want 7", got)
	}
}

func TestClosureCounter(t *testing.T) {
	code := `
	fun Counter() {
		count := 0
		fun inc() { count = count + 1  return count }
		return this
	}
	c := Counter()
	c.inc()
	c.inc()
	c.inc()`
	if got := run(t, code); got != 3.0 {
		t.Errorf("counter = %v, want 3", got)
	}
}

func TestNamedFunctionInTable(t *testing.T) {
	code := `
	fun square(x) { return x * x }
	t := { f: square }
	t.f(6)`
	if got := run(t, code); got != 36.0 {
		t.Errorf("t.f(6) = %v, want 36", got)
	}
}

func TestAnonFunctionAsMethod(t *testing.T) {
	code := `
	dog := {
		name: "旺财",
		hello: fun() { return this.name }
	}
	dog.hello()`
	if got := run(t, code); got != "旺财" {
		t.Errorf("dog.hello() = %v, want 旺财", got)
	}
}

// TestThisReceiverMutation 方法通过 this 读写字段，改动落在原对象上
func TestThisReceiverMutation(t *testing.T) {
	code := `
	fun newAccount(balance) {
		return {
			balance: balance,
			deposit: fun(n) { this.balance = this.balance + n  return this.balance }
		}
	}
	acc := newAccount(100)
	acc.deposit(50)
	acc.balance`
	if got := run(t, code); got != 150.0 {
		t.Errorf("acc.balance = %v, want 150", got)
	}
}

func TestObjectInstancesAreIndependent(t *testing.T) {
	code := `
	fun box(v) {
		return { v: v, set: fun(x) { this.v = x }, get: fun() { return this.v } }
	}
	a := box(1)
	b := box(2)
	a.set(100)
	a.get() + b.get()`
	if got := run(t, code); got != 102.0 {
		t.Errorf("a.get() + b.get() = %v, want 102", got)
	}
}

func TestHigherOrderFunction(t *testing.T) {
	code := `
	fun apply(f, x) { return f(x) }
	apply(fun(n) { return n + 1 }, 41)`
	if got := run(t, code); got != 42.0 {
		t.Errorf("apply = %v, want 42", got)
	}
}

func TestElseIfChain(t *testing.T) {
	code := `
	fun grade(s) {
		if s >= 90 { return "A" }
		else if s >= 60 { return "B" }
		else { return "C" }
	}
	grade(75)`
	if got := run(t, code); got != "B" {
		t.Errorf("grade(75) = %v, want B", got)
	}
}

func TestBuiltins(t *testing.T) {
	if got := run(t, `len("hello")`); got != 5 {
		t.Errorf("len = %v, want 5", got)
	}
	if got := run(t, `type({ a: 1 })`); got != "table" {
		t.Errorf("type = %v, want table", got)
	}
	if got := run(t, `num("42") + 8`); got != 50.0 {
		t.Errorf("num = %v, want 50", got)
	}
	if got := run(t, `upper("fn")`); got != "FN" {
		t.Errorf("upper = %v, want FN", got)
	}
}

func TestStringIndex(t *testing.T) {
	if got := run(t, `s := "hello"  s[1]`); got != "e" {
		t.Errorf("s[1] = %v, want e", got)
	}
}

// TestOperatorNameNotShadowed 用户函数命名为 add 不应覆盖 + 运算
func TestOperatorNameNotShadowed(t *testing.T) {
	code := `
	fun add(a, b) { return a + b + 1 }
	add(2, 3) + (2 + 3)`
	if got := run(t, code); got != 11.0 { // (2+3+1) + (2+3) = 6 + 5
		t.Errorf("got %v, want 11", got)
	}
}

func TestCoroutineYield(t *testing.T) {
	code := `
	fun gen() {
		yield(1)
		yield(2)
		return 3
	}
	co := coroutine(gen)
	a := resume(co)
	b := resume(co)
	c := resume(co)
	a + b * 10 + c * 100`
	if got := run(t, code); got != 321.0 {
		t.Errorf("coroutine yields = %v, want 321", got)
	}
}

func TestCoroutineStatus(t *testing.T) {
	code := `
	fun gen() { yield(1) }
	co := coroutine(gen)
	before := costatus(co)
	resume(co)
	mid := costatus(co)
	resume(co)
	after := costatus(co)
	before + "," + mid + "," + after`
	if got := run(t, code); got != "suspended,suspended,dead" {
		t.Errorf("status = %v, want suspended,suspended,dead", got)
	}
}

func TestCoroutineTwoWay(t *testing.T) {
	// yield 的返回值来自下一次 resume 的参数
	code := `
	fun adder() {
		x := yield(0)
		y := yield(x)
		return x + y
	}
	co := coroutine(adder)
	resume(co)          // 跑到第一个 yield
	resume(co, 10)      // x = 10, 跑到第二个 yield
	resume(co, 20)      // y = 20, 返回 x + y`
	if got := run(t, code); got != 30.0 {
		t.Errorf("two-way coroutine = %v, want 30", got)
	}
}

func TestCoroutineType(t *testing.T) {
	code := `
	fun g() { yield(1) }
	type(coroutine(g))`
	if got := run(t, code); got != "coroutine" {
		t.Errorf("type = %v, want coroutine", got)
	}
}

// ---------- 管道 | ----------

func TestPipeBasic(t *testing.T) {
	code := `
	fun double(x) { return x * 2 }
	fun plus(a, b) { return a + b }
	5 | double | plus(3)` // (5*2)+3
	if got := run(t, code); got != 13.0 {
		t.Errorf("pipe = %v, want 13", got)
	}
}

func TestPipeIntoBuiltin(t *testing.T) {
	if got := run(t, `"fun" | upper`); got != "FUN" {
		t.Errorf(`"fun" | upper = %v, want FUN`, got)
	}
}

func TestPipeIntoMethod(t *testing.T) {
	code := `
	obj := { wrap: fun(s) { return "[" + s + "]" } }
	"x" | obj.wrap`
	if got := run(t, code); got != "[x]" {
		t.Errorf("pipe into method = %v, want [x]", got)
	}
}

// ---------- 值调用（下标/表达式结果直接调用）----------

func TestCallValueFromIndex(t *testing.T) {
	code := `
	handlers := { hi: fun(n) { return "hi " + n } }
	cmd := "hi"
	handlers[cmd]("gin")`
	if got := run(t, code); got != "hi gin" {
		t.Errorf("handlers[cmd](x) = %v, want hi gin", got)
	}
}

func TestCallValueCurried(t *testing.T) {
	code := `
	fun adder(a) { return fun(b) { return a + b } }
	adder(3)(4)`
	if got := run(t, code); got != 7.0 {
		t.Errorf("adder(3)(4) = %v, want 7", got)
	}
}

// ---------- 完整 OOP：clone / merge / 继承 / 多态 / super ----------

func TestArrayLiteral(t *testing.T) {
	code := `arr := { 10, 20, 30 }  arr[0] + arr[2] * len(arr)`
	if got := run(t, code); got != 100.0 { // 10 + 30*3
		t.Errorf("array literal = %v, want 100", got)
	}
}

func TestCloneIsIndependent(t *testing.T) {
	code := `
	proto := { v: 1 }
	a := clone(proto)
	a.v = 99
	proto.v`
	if got := run(t, code); got != 1 {
		t.Errorf("proto.v = %v, want 1 (clone 应独立)", got)
	}
}

func TestInheritanceOverride(t *testing.T) {
	code := `
	Animal := {
		init:  fun(n) { this.name = n  return this },
		speak: fun() { return this.name + ": ..." }
	}
	Dog := merge(clone(Animal), {
		speak: fun() { return this.name + ": 汪" }
	})
	clone(Dog).init("旺财").speak()`
	if got := run(t, code); got != "旺财: 汪" {
		t.Errorf("override = %v, want 旺财: 汪", got)
	}
}

func TestSuperLateBinding(t *testing.T) {
	// 父类方法挂到子类字段上，this 依旧晚绑定到实例
	code := `
	Animal := {
		init:  fun(n) { this.name = n  return this },
		speak: fun() { return this.name + ": base" }
	}
	Dog := merge(clone(Animal), {
		superSpeak: Animal.speak,
		speak: fun() { return this.superSpeak() + "+dog" }
	})
	clone(Dog).init("d").speak()`
	if got := run(t, code); got != "d: base+dog" {
		t.Errorf("super = %v, want d: base+dog", got)
	}
}

func TestMixin(t *testing.T) {
	code := `
	A := { a: fun() { return 1 } }
	B := { b: fun() { return 2 } }
	C := merge(clone(A), B)
	C.a() + C.b()`
	if got := run(t, code); got != 3.0 {
		t.Errorf("mixin = %v, want 3", got)
	}
}

func TestMethodChainingLong(t *testing.T) {
	code := `
	fun box() {
		return { v: 0, add: fun(n) { this.v = this.v + n  return this } }
	}
	box().add(1).add(2).add(3).v`
	if got := run(t, code); got != 6.0 {
		t.Errorf("chain = %v, want 6", got)
	}
}

// runInDir 在指定目录下求值代码（供 import 测试解析相对路径）
func runInDir(t *testing.T, dir, code string) any {
	t.Helper()
	tokens := lexer.NewLexer().Tokenize(code)
	ast := parser.NewParser().Parse(tokens)
	machine := vm.NewVM()
	machine.SetDir(dir)
	result, err := machine.Call(ast)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	return result
}

func TestModuleImplicitExport(t *testing.T) {
	dir := t.TempDir()
	// 模块：顶层定义的名字都被导出（add 命名也不会破坏 + 运算）
	mod := "PI := 3\nfun add(a, b) { return a + b }\n"
	if err := os.WriteFile(filepath.Join(dir, "math.fun"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}
	code := `m := import("math.fun")  m.add(m.PI, 4)`
	if got := runInDir(t, dir, code); got != 7.0 {
		t.Errorf("m.add(m.PI, 4) = %v, want 7", got)
	}
}

func TestModuleExplicitExport(t *testing.T) {
	dir := t.TempDir()
	// 显式导出：只暴露 return 的表，secret 不外泄
	mod := "secret := 42\nfun pub() { return 1 }\nreturn { pub: pub }\n"
	if err := os.WriteFile(filepath.Join(dir, "m.fun"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}
	code := `m := import("m")  has(m, "secret")`
	if got := runInDir(t, dir, code); got != false {
		t.Errorf("has(m, secret) = %v, want false", got)
	}
	code2 := `m := import("m")  m.pub()`
	if got := runInDir(t, dir, code2); got != 1 { // return 1 是整数字面量
		t.Errorf("m.pub() = %v, want 1", got)
	}
}
