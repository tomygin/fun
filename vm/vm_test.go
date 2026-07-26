package vm_test

import (
	"fun/lexer"
	"fun/parser"
	"fun/vm"
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
