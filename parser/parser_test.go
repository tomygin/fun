package parser_test

import (
	"fmt"
	"fun/lexer"
	"fun/parser"
	"testing"
)

// parse 返回整个程序的 s-expression（["begin", ...]）
func parse(t *testing.T, code string) []any {
	t.Helper()
	tokens := lexer.NewLexer().Tokenize(code)
	return parser.NewParser().Parse(tokens)
}

// first 返回第一条语句的 s-expression 字符串形态
func first(t *testing.T, code string) string {
	t.Helper()
	ast := parse(t, code)
	if len(ast) < 2 {
		t.Fatalf("%q: 没有语句", code)
	}
	return fmt.Sprint(ast[1])
}

func TestOperatorPrecedence(t *testing.T) {
	// 乘法优先于加法
	if got := first(t, `1 + 2 * 3`); got != "[@add 1 [@mul 2 3]]" {
		t.Errorf("1+2*3 => %s", got)
	}
	// 括号改变结合
	if got := first(t, `(1 + 2) * 3`); got != "[@mul [@add 1 2] 3]" {
		t.Errorf("(1+2)*3 => %s", got)
	}
	// 比较低于加法
	if got := first(t, `1 + 2 > 2`); got != "[@gt [@add 1 2] 2]" {
		t.Errorf("1+2>2 => %s", got)
	}
	// && 低于比较，|| 低于 &&
	if got := first(t, `a > b && c || d`); got != "[or [and [@gt a b] c] d]" {
		t.Errorf("logic => %s", got)
	}
}

func TestUnaryOperators(t *testing.T) {
	if got := first(t, `-x`); got != "[@sub 0 x]" {
		t.Errorf("-x => %s", got)
	}
	if got := first(t, `!ok`); got != "[not ok]" {
		t.Errorf("!ok => %s", got)
	}
}

func TestPipeRewrite(t *testing.T) {
	// x | f => f(x)
	if got := first(t, `x | f`); got != "[f x]" {
		t.Errorf("x|f => %s", got)
	}
	// 左值插为第一实参
	if got := first(t, `x | g(1)`); got != "[g x 1]" {
		t.Errorf("x|g(1) => %s", got)
	}
	// 链式左结合
	if got := first(t, `x | f | g`); got != "[g [f x]]" {
		t.Errorf("x|f|g => %s", got)
	}
	// 管道进方法：this 绑定 obj
	if got := first(t, `x | obj.m`); got != "[method-call obj m x]" {
		t.Errorf("x|obj.m => %s", got)
	}
}

func TestForDesugarsToWhile(t *testing.T) {
	// 条件式 for -> while 节点
	if got := first(t, `for a < 3 { x = 1 }`); got != "[while [@lt a 3] [begin [assign x 1]]]" {
		t.Errorf("for-cond => %s", got)
	}
	// 无限 for -> while true
	if got := first(t, `for { x = 1 }`); got != "[while true [begin [assign x 1]]]" {
		t.Errorf("for-inf => %s", got)
	}
	// 三段式脱糖为 begin{init, while{body,post}}
	got := first(t, `for i := 0; i < 3; i++ { x = i }`)
	want := "[begin [var i 0] [while [@lt i 3] [begin [assign x i] [assign i [@add i 1]]]]]"
	if got != want {
		t.Errorf("for-3 => %s", got)
	}
}

func TestSelfIncrementSugar(t *testing.T) {
	if got := first(t, `i++`); got != "[assign i [@add i 1]]" {
		t.Errorf("i++ => %s", got)
	}
}

func TestAssignTargets(t *testing.T) {
	if got := first(t, `a = 1`); got != "[assign a 1]" {
		t.Errorf("a=1 => %s", got)
	}
	if got := first(t, `a.b = 1`); got != "[assign [property a b] 1]" {
		t.Errorf("a.b=1 => %s", got)
	}
	if got := first(t, `a[k] = 1`); got != "[assign [index a k] 1]" {
		t.Errorf("a[k]=1 => %s", got)
	}
	if got := first(t, `this.x = 1`); got != "[assign [property this x] 1]" {
		t.Errorf("this.x=1 => %s", got)
	}
}

func TestTableLiteral(t *testing.T) {
	// 键值对：键被包成字符串常量
	if got := first(t, `{ a: 1 }`); got != `[table "a" 1]` {
		t.Errorf("{a:1} => %s", got)
	}
	// 数组式自动编号
	if got := first(t, `{ 7, 8 }`); got != `[table "0" 7 "1" 8]` {
		t.Errorf("{7,8} => %s", got)
	}
	// 混写
	if got := first(t, `{ 7, a: 1 }`); got != `[table "0" 7 "a" 1]` {
		t.Errorf("{7,a:1} => %s", got)
	}
	// 整数键
	if got := first(t, `{ 0: "x" }`); got != `[table "0" "x"]` {
		t.Errorf("{0:x} => %s", got)
	}
}

func TestElseIfChain(t *testing.T) {
	// else if 变成 else 分支里的嵌套 if
	got := first(t, `if a { x = 1 } else if b { x = 2 }`)
	want := "[if a [begin [assign x 1]] [begin [if b [begin [assign x 2]] [begin]]]]"
	if got != want {
		t.Errorf("else-if => %s", got)
	}
}

func TestAnonymousFunction(t *testing.T) {
	if got := first(t, `fun(x) { return x }`); got != "[fun-expr [x] [begin [return x]]]" {
		t.Errorf("fun-expr => %s", got)
	}
}

func TestPostfixChain(t *testing.T) {
	// 属性 -> 方法 -> 下标 -> 调用括号
	if got := first(t, `a.b.c`); got != "[property [property a b] c]" {
		t.Errorf("a.b.c => %s", got)
	}
	if got := first(t, `t[k](1)`); got != "[call-value [index t k] 1]" {
		t.Errorf("t[k](1) => %s", got)
	}
	if got := first(t, `f(1)(2)`); got != "[call-value [f 1] 2]" {
		t.Errorf("f(1)(2) => %s", got)
	}
}

func TestCallValueRequiresSameLine(t *testing.T) {
	// 换行的括号不是调用：a 和 (b) 是两条语句
	ast := parse(t, "a\n(b)")
	if len(ast) != 3 { // begin + 2 条语句
		t.Errorf("跨行括号被误吞成调用: %v", fmt.Sprint(ast))
	}
}

func TestPipeLowestPrecedence(t *testing.T) {
	// 管道低于 ||：a || b | f  =>  f(a || b)
	if got := first(t, `a || b | f`); got != "[f [or a b]]" {
		t.Errorf("pipe vs || => %s", got)
	}
	// 管道右侧的 @ 运算符调用
	if got := first(t, `x | @add(3)`); got != "[@add x 3]" {
		t.Errorf("pipe into @op => %s", got)
	}
}

func TestStatementLevelAnonymousFun(t *testing.T) {
	// 语句位置的匿名函数是表达式，不是具名定义
	if got := first(t, `fun(x) { return x }`); got != "[fun-expr [x] [begin [return x]]]" {
		t.Errorf("statement fun-expr => %s", got)
	}
	// 具名定义不受影响
	if got := first(t, `fun f(x) { return x }`); got != "[fun f [x] [begin [return x]]]" {
		t.Errorf("named fun => %s", got)
	}
}

func TestNestedTableLiteral(t *testing.T) {
	if got := first(t, `{ a: { b: 1 } }`); got != `[table "a" [table "b" 1]]` {
		t.Errorf("nested table => %s", got)
	}
	// 表里放匿名函数
	if got := first(t, `{ f: fun() { return 1 } }`); got != `[table "f" [fun-expr [] [begin [return 1]]]]` {
		t.Errorf("table with fun => %s", got)
	}
}

func TestMethodCallWithArgs(t *testing.T) {
	if got := first(t, `obj.m(1, 2)`); got != "[method-call obj m 1 2]" {
		t.Errorf("method args => %s", got)
	}
}

func TestReturnWithoutValue(t *testing.T) {
	got := first(t, `fun f() { return }`)
	if got != "[fun f [] [begin [return <nil>]]]" {
		t.Errorf("bare return => %s", got)
	}
}

func TestOperatorsCompileToAtSpace(t *testing.T) {
	// 全部运算符落入 @ 空间
	ops := map[string]string{
		`a + b`: "[@add a b]", `a - b`: "[@sub a b]",
		`a * b`: "[@mul a b]", `a / b`: "[@div a b]",
		`a % b`: "[@mod a b]", `a == b`: "[@eq a b]",
		`a != b`: "[@neq a b]", `a > b`: "[@gt a b]",
		`a < b`: "[@lt a b]", `a >= b`: "[@gte a b]", `a <= b`: "[@lte a b]",
	}
	for code, want := range ops {
		if got := first(t, code); got != want {
			t.Errorf("%s => %s, want %s", code, got, want)
		}
	}
}
