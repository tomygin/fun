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
	if got := run(t, `1 + 2 * 3`); got != 7 {
		t.Errorf("1 + 2 * 3 = %v, want 7", got)
	}
	if got := run(t, `10 % 3`); got != 1 {
		t.Errorf("10 %% 3 = %v, want 1", got)
	}
	if got := run(t, `-5 + 2`); got != -3 {
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
	if got := run(t, code); got != 55 {
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
	if got := run(t, code); got != 6 {
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
	if got := run(t, code); got != 16 {
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
	if got := run(t, code); got != 128 {
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
	if got := run(t, code); got != 5050 {
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
	if got := run(t, code); got != 7 {
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
	if got := run(t, code); got != 3 {
		t.Errorf("counter = %v, want 3", got)
	}
}

func TestNamedFunctionInTable(t *testing.T) {
	code := `
	fun square(x) { return x * x }
	t := { f: square }
	t.f(6)`
	if got := run(t, code); got != 36 {
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
	if got := run(t, code); got != 150 {
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
	if got := run(t, code); got != 102 {
		t.Errorf("a.get() + b.get() = %v, want 102", got)
	}
}

func TestHigherOrderFunction(t *testing.T) {
	code := `
	fun apply(f, x) { return f(x) }
	apply(fun(n) { return n + 1 }, 41)`
	if got := run(t, code); got != 42 {
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
	if got := run(t, `num("42") + 8`); got != 50 {
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
	if got := run(t, code); got != 11 { // (2+3+1) + (2+3) = 6 + 5
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
	if got := run(t, code); got != 321 {
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
	if got := run(t, code); got != 30 {
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
	if got := run(t, code); got != 13 {
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
	if got := run(t, code); got != 7 {
		t.Errorf("adder(3)(4) = %v, want 7", got)
	}
}

// ---------- 完整 OOP：clone / merge / 继承 / 多态 / super ----------

func TestArrayLiteral(t *testing.T) {
	code := `arr := { 10, 20, 30 }  arr[0] + arr[2] * len(arr)`
	if got := run(t, code); got != 100 { // 10 + 30*3
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
	if got := run(t, code); got != 3 {
		t.Errorf("mixin = %v, want 3", got)
	}
}

func TestMethodChainingLong(t *testing.T) {
	code := `
	fun box() {
		return { v: 0, add: fun(n) { this.v = this.v + n  return this } }
	}
	box().add(1).add(2).add(3).v`
	if got := run(t, code); got != 6 {
		t.Errorf("chain = %v, want 6", got)
	}
}

// ---------- 元编程 @ 空间 ----------

func TestOperatorLocalOverride(t *testing.T) {
	// 函数内覆盖 @add，退出后自动恢复
	code := `
	fun weird() {
		@add := fun(a, b) { return 100 }
		return 1 + 2
	}
	w := weird()
	n := 1 + 2
	w * 1000 + n` // 100*1000 + 3
	if got := run(t, code); got != 100003 {
		t.Errorf("override = %v, want 100003", got)
	}
}

func TestOperatorAsValue(t *testing.T) {
	// 运算符本身是一等值，可以当参数传
	code := `
	fun reduce(t, f, init) {
		acc := init
		for i := 0; i < len(t); i++ { acc = f(acc, t[i]) }
		return acc
	}
	reduce({ 1, 2, 3, 4 }, @add, 0)`
	if got := run(t, code); got != 10 {
		t.Errorf("reduce(@add) = %v, want 10", got)
	}
}

func TestOperatorSaveAndWrap(t *testing.T) {
	// save-and-wrap：覆盖体内用保存的原实现
	code := `
	fun demo() {
		old := @add
		@add := fun(a, b) { return old(a, b) * 10 }
		return 1 + 2
	}
	demo()`
	if got := run(t, code); got != 30 {
		t.Errorf("save-and-wrap = %v, want 30", got)
	}
}

// ---------- 字符串：转义 / 模板 ----------

func TestStringEscapes(t *testing.T) {
	if got := run(t, `"a\nb"`); got != "a\nb" {
		t.Errorf("escape \\n = %q", got)
	}
	if got := run(t, `"say \"hi\""`); got != `say "hi"` {
		t.Errorf("escape quote = %q", got)
	}
	if got := run(t, `'it\'s'`); got != "it's" {
		t.Errorf("escape single quote = %q", got)
	}
	if got := run(t, `"back\\slash"`); got != `back\slash` {
		t.Errorf("escape backslash = %q", got)
	}
}

func TestTemplateString(t *testing.T) {
	code := "name := \"Gin\"\n`hi ${name}, ${1 + 2}`"
	if got := run(t, code); got != "hi Gin, 3" {
		t.Errorf("template = %q, want hi Gin, 3", got)
	}
}

func TestTemplateStringRaw(t *testing.T) {
	// 反引号内不处理转义，\n 原样保留
	code := "`a\\nb`"
	if got := run(t, code); got != `a\nb` {
		t.Errorf("raw = %q, want a\\nb", got)
	}
}

func TestTemplateStringMultiline(t *testing.T) {
	code := "`line1\nline2`"
	if got := run(t, code); got != "line1\nline2" {
		t.Errorf("multiline = %q", got)
	}
}

func TestTemplateStringNestedBraces(t *testing.T) {
	code := "`n=${len({ 1, 2, 3 })}`"
	if got := run(t, code); got != "n=3" {
		t.Errorf("nested braces = %q, want n=3", got)
	}
}

// ---------- 真值规则 ----------

func TestFalsyValues(t *testing.T) {
	falsy := []string{`bool(false)`, `bool(nil)`, `bool(0)`, `bool(0.0)`, `bool("")`, `bool({})`}
	for _, code := range falsy {
		if got := run(t, code); got != false {
			t.Errorf("%s = %v, want false", code, got)
		}
	}
	truthy := []string{`bool("0")`, `bool("false")`, `bool(-1)`, `bool({ 1 })`}
	for _, code := range truthy {
		if got := run(t, code); got != true {
			t.Errorf("%s = %v, want true", code, got)
		}
	}
}

// ---------- 数字塔：精度与溢出 ----------

func TestDecimalExact(t *testing.T) {
	if got := run(t, `0.1 + 0.2 == 0.3`); got != true {
		t.Errorf("0.1 + 0.2 == 0.3 => %v, want true", got)
	}
	if got := run(t, `str(0.1 + 0.2)`); got != "0.3" {
		t.Errorf("str(0.1+0.2) = %q, want 0.3", got)
	}
	if got := run(t, `0.1 * 0.2 == 0.02`); got != true {
		t.Errorf("0.1 * 0.2 == 0.02 => %v", got)
	}
	if got := run(t, `str(1.1 - 1)`); got != "0.1" {
		t.Errorf("str(1.1-1) = %q, want 0.1", got)
	}
}

func TestIntNoPrecisionLoss(t *testing.T) {
	// float64 的 53 位尾数问题：2^53 + 1 必须精确
	if got := run(t, `str(9007199254740992 + 1)`); got != "9007199254740993" {
		t.Errorf("2^53+1 = %v", got)
	}
}

func TestIntOverflowPromotion(t *testing.T) {
	// int64 溢出自动提升为大数，永不回绕
	if got := run(t, `str(9223372036854775807 + 1)`); got != "9223372036854775808" {
		t.Errorf("int64max+1 = %v", got)
	}
	if got := run(t, `9223372036854775807 + 1 > 9223372036854775807`); got != true {
		t.Errorf("大数比较错误")
	}
	if got := run(t, `9223372036854775807 + 1 - 1 == 9223372036854775807`); got != true {
		t.Errorf("大数运算往返错误")
	}
}

func TestExactDivision(t *testing.T) {
	if got := run(t, `1 / 4 == 0.25`); got != true {
		t.Errorf("1/4 == 0.25 => %v", got)
	}
	if got := run(t, `10 / 2`); got != 5 {
		t.Errorf("10/2 = %v, want 5", got)
	}
	// 1/3 内部是精确有理数：乘回去等于 1
	if got := run(t, `1 / 3 * 3 == 1`); got != true {
		t.Errorf("1/3*3 == 1 => %v, want true（内部精确）", got)
	}
	// 无限循环小数打印截断到 20 位
	if got := run(t, `str(1 / 3)`); got != "0.33333333333333333333" {
		t.Errorf("str(1/3) = %q", got)
	}
}

func TestDecimalMod(t *testing.T) {
	if got := run(t, `str(5.5 % 2)`); got != "1.5" {
		t.Errorf("5.5 %% 2 = %v, want 1.5", got)
	}
}

func TestNumberNormalization(t *testing.T) {
	// 0.5+0.5 归一化降回整数；2.50 归一为 2.5
	if got := run(t, `0.5 + 0.5`); got != 1 {
		t.Errorf("0.5+0.5 = %v, want int 1", got)
	}
	if got := run(t, `str(2.50)`); got != "2.5" {
		t.Errorf("str(2.50) = %q, want 2.5", got)
	}
	if got := run(t, `1 == 1.0`); got != true {
		t.Errorf("1 == 1.0 => %v", got)
	}
}

func TestNumParsesExactDecimal(t *testing.T) {
	if got := run(t, `num("0.1") + num("0.2") == 0.3`); got != true {
		t.Errorf(`num("0.1")+num("0.2") == 0.3 => %v`, got)
	}
}

// ---------- merge 优先级 ----------

func TestMergePriority(t *testing.T) {
	// 越靠后的表优先级越高
	code := `merge({ a: 1, b: 1 }, { a: 2 }, { a: 3, b: 2 }).a`
	if got := run(t, code); got != 3 {
		t.Errorf("merge priority = %v, want 3 (最后的表赢)", got)
	}
}

// ---------- JSON ----------

func TestJSONRoundTrip(t *testing.T) {
	code := `
	s := json.encode({ name: "Gin", age: 18, tags: { "a", "b" } })
	back := json.decode(s)
	back.name + str(back.age) + back.tags[1]`
	if got := run(t, code); got != "Gin18b" {
		t.Errorf("json roundtrip = %v, want Gin18b", got)
	}
}

func TestJSONArrayEncoding(t *testing.T) {
	if got := run(t, `json.encode({ 1, 2, 3 })`); got != "[1,2,3]" {
		t.Errorf("array encode = %v, want [1,2,3]", got)
	}
}

func TestJSONExactDecimal(t *testing.T) {
	// JSON 数字走数字塔：0.1 解码后依然精确
	code := `back := json.decode('{"x": 0.1}')  back.x + 0.2 == 0.3`
	if got := run(t, code); got != true {
		t.Errorf("json decimal exact = %v, want true", got)
	}
}

// ---------- 错误处理：throw / try ----------

func TestTryOk(t *testing.T) {
	code := `
	r := try(fun() { return 42 })
	str(r.ok) + "/" + str(r.value)`
	if got := run(t, code); got != "true/42" {
		t.Errorf("try ok = %v, want true/42", got)
	}
}

func TestThrowAndCatch(t *testing.T) {
	code := `
	fun withdraw(balance, n) {
		if n > balance { throw("余额不足") }
		return balance - n
	}
	r := try(withdraw, 100, 500)
	str(r.ok) + "/" + r.error`
	if got := run(t, code); got != "false/余额不足" {
		t.Errorf("throw/catch = %v", got)
	}
}

func TestThrowTableError(t *testing.T) {
	code := `
	r := try(fun() { throw({ code: 404, msg: "not found" }) })
	str(r.error.code) + "/" + r.error.msg`
	if got := run(t, code); got != "404/not found" {
		t.Errorf("table error = %v", got)
	}
}

func TestTryCatchesRuntimeError(t *testing.T) {
	code := `
	r := try(fun() { return no_such_variable })
	r.ok`
	if got := run(t, code); got != false {
		t.Errorf("runtime error not caught: %v", got)
	}
}

func TestErrorPropagatesThroughCalls(t *testing.T) {
	code := `
	fun deep() { throw("deep error") }
	fun mid() { deep()  return "unreachable" }
	r := try(mid)
	r.error`
	if got := run(t, code); got != "deep error" {
		t.Errorf("propagation = %v", got)
	}
}

func TestNestedTry(t *testing.T) {
	code := `
	r := try(fun() {
		inner := try(fun() { throw("inner") })
		return "caught: " + inner.error
	})
	str(r.ok) + "/" + r.value`
	if got := run(t, code); got != "true/caught: inner" {
		t.Errorf("nested try = %v", got)
	}
}

func TestAssertCatchable(t *testing.T) {
	code := `
	r := try(fun() { assert(1 > 2, "impossible") })
	str(r.ok) + "/" + r.error`
	if got := run(t, code); got != "false/impossible" {
		t.Errorf("assert catchable = %v", got)
	}
}

func TestUncaughtThrowBubbles(t *testing.T) {
	tokens := lexer.NewLexer().Tokenize(`throw("boom")`)
	ast := parser.NewParser().Parse(tokens)
	_, err := vm.NewVM().Call(ast)
	fe, ok := err.(*vm.FunError)
	if !ok || fe.Value != "boom" {
		t.Errorf("uncaught throw = %v, want FunError(boom)", err)
	}
}

// ---------- JSON 补齐：错误与美化 ----------

func TestJSONDecodeInvalidCatchable(t *testing.T) {
	code := `
	r := try(json.decode, "{bad")
	r.ok`
	if got := run(t, code); got != false {
		t.Errorf("invalid json not caught: %v", got)
	}
}

func TestJSONDecodeNullVsError(t *testing.T) {
	// 合法的 "null" 解码成 nil，与解析失败可区分
	code := `
	r := try(json.decode, "null")
	str(r.ok) + "/" + str(r.value)`
	if got := run(t, code); got != "true/nil" {
		t.Errorf("null decode = %v, want true/nil", got)
	}
}

func TestJSONEncodeIndent(t *testing.T) {
	code := `json.encode({ a: 1 }, 2)`
	if got := run(t, code); got != "{\n  \"a\": 1\n}" {
		t.Errorf("indent = %q", got)
	}
}

// ---------- HTTP：服务端 + 客户端 ----------

func TestHTTPServerAndClient(t *testing.T) {
	code := `
	routes := {
		"/hello": fun(req) { return "hi, " + req.query["name"] },
		"/echo":  fun(req) { return { status: 201, body: req.body } },
		"*":      fun(req) { return { status: 404, body: "miss" } }
	}
	http.listen("127.0.0.1:18921", routes)
	r1 := http.get("http://127.0.0.1:18921/hello?name=Tom")
	r2 := http.post("http://127.0.0.1:18921/echo", "ping")
	r3 := http.get("http://127.0.0.1:18921/nope")
	r1.body + "|" + str(r2.status) + r2.body + "|" + str(r3.status) + str(r3.ok)`
	if got := run(t, code); got != "hi, Tom|201ping|404false" {
		t.Errorf("http roundtrip = %v", got)
	}
}

func TestHTTPHandlerUsesInterpreterState(t *testing.T) {
	// 处理函数是真正的 Fun 闭包：能读写脚本里的变量
	code := `
	count := 0
	http.listen("127.0.0.1:18922", {
		"/inc": fun(req) { count = count + 1  return str(count) }
	})
	http.get("http://127.0.0.1:18922/inc")
	http.get("http://127.0.0.1:18922/inc")
	r := http.get("http://127.0.0.1:18922/inc")
	r.body + "/" + str(count)`
	if got := run(t, code); got != "3/3" {
		t.Errorf("handler closure = %v, want 3/3", got)
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
	if got := runInDir(t, dir, code); got != 7 {
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
