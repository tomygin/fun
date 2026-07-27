package number_test

import (
	"fun/number"
	"math/big"
	"testing"
)

func TestParseInt(t *testing.T) {
	v, err := number.ParseInt("42")
	if err != nil || v != 42 {
		t.Errorf("ParseInt(42) = %v, %v", v, err)
	}
	// 超出 int64：自动大数，不报错
	v, err = number.ParseInt("99999999999999999999999999")
	if err != nil {
		t.Fatalf("大整数字面量报错: %v", err)
	}
	if _, ok := v.(*big.Rat); !ok {
		t.Errorf("大整数应是 *big.Rat, got %T", v)
	}
	if number.Format(v) != "99999999999999999999999999" {
		t.Errorf("大整数打印 = %s", number.Format(v))
	}
}

func TestParseDec(t *testing.T) {
	v, err := number.ParseDec("0.1")
	if err != nil {
		t.Fatal(err)
	}
	if d, ok := v.(number.Dec); !ok || d.Coef != 1 || d.Scale != 1 {
		t.Errorf("0.1 = %#v, want Dec{1,1}", v)
	}
	// 尾零归一化
	v, _ = number.ParseDec("2.50")
	if d, ok := v.(number.Dec); !ok || d.Coef != 25 || d.Scale != 1 {
		t.Errorf("2.50 = %#v, want Dec{25,1}", v)
	}
	// 小数位全是零：降级成 int
	v, _ = number.ParseDec("3.0")
	if v != 3 {
		t.Errorf("3.0 = %#v, want int 3", v)
	}
	// 负小数
	v, _ = number.ParseDec("-0.5")
	if d, ok := v.(number.Dec); !ok || d.Coef != -5 || d.Scale != 1 {
		t.Errorf("-0.5 = %#v, want Dec{-5,1}", v)
	}
}

func TestAddExactDecimal(t *testing.T) {
	a, _ := number.ParseDec("0.1")
	b, _ := number.ParseDec("0.2")
	sum, ok := number.Add(a, b)
	if !ok {
		t.Fatal("Add failed")
	}
	want, _ := number.ParseDec("0.3")
	if c, _ := number.Cmp(sum, want); c != 0 {
		t.Errorf("0.1+0.2 = %v, want 0.3", number.Format(sum))
	}
	if number.Format(sum) != "0.3" {
		t.Errorf("Format(0.1+0.2) = %s", number.Format(sum))
	}
}

func TestIntOverflowPromotes(t *testing.T) {
	max := int(9223372036854775807)
	sum, ok := number.Add(max, 1)
	if !ok {
		t.Fatal("Add failed")
	}
	if number.Format(sum) != "9223372036854775808" {
		t.Errorf("int64max+1 = %s", number.Format(sum))
	}
	// 溢出后减回来能降回 int
	back, _ := number.Sub(sum, 1)
	if back != max {
		t.Errorf("提升后降级失败: %v (%T)", back, back)
	}
}

func TestMulOverflowPromotes(t *testing.T) {
	big1 := int(3037000500) // sqrt(int64max) 附近
	p, ok := number.Mul(big1, big1)
	if !ok {
		t.Fatal("Mul failed")
	}
	if number.Format(p) != "9223372037000250000" {
		t.Errorf("溢出乘法 = %s", number.Format(p))
	}
}

func TestDivExactAndInfinite(t *testing.T) {
	// 整除 -> int
	q, _ := number.Div(10, 2)
	if q != 5 {
		t.Errorf("10/2 = %v (%T)", q, q)
	}
	// 有限小数 -> Dec
	q, _ = number.Div(1, 4)
	if d, ok := q.(number.Dec); !ok || d.Coef != 25 || d.Scale != 2 {
		t.Errorf("1/4 = %#v, want Dec{25,2}", q)
	}
	// 无限循环 -> Rat，且乘回去精确
	q, _ = number.Div(1, 3)
	if _, ok := q.(*big.Rat); !ok {
		t.Errorf("1/3 应是 *big.Rat, got %T", q)
	}
	p, _ := number.Mul(q, 3)
	if p != 1 {
		t.Errorf("1/3*3 = %v, want int 1", p)
	}
}

func TestFormatRatTruncation(t *testing.T) {
	q, _ := number.Div(1, 3)
	if number.Format(q) != "0.33333333333333333333" {
		t.Errorf("Format(1/3) = %s", number.Format(q))
	}
	// 负数定点格式
	d, _ := number.ParseDec("-0.05")
	if number.Format(d) != "-0.05" {
		t.Errorf("Format(-0.05) = %s", number.Format(d))
	}
}

func TestCmpCrossTier(t *testing.T) {
	// int vs Dec
	half, _ := number.ParseDec("0.5")
	if c, _ := number.Cmp(1, half); c != 1 {
		t.Errorf("1 vs 0.5 = %d", c)
	}
	// int vs 提升后的大数
	huge, _ := number.Add(int(9223372036854775807), 1)
	if c, _ := number.Cmp(huge, 100); c != 1 {
		t.Errorf("huge vs 100 = %d", c)
	}
	// 相等跨层
	one, _ := number.ParseDec("1.0") // 归一化成 int 1
	if c, _ := number.Cmp(1, one); c != 0 {
		t.Errorf("1 vs 1.0 = %d", c)
	}
}

func TestModSemantics(t *testing.T) {
	m, _ := number.Mod(10, 3)
	if m != 1 {
		t.Errorf("10%%3 = %v", m)
	}
	// 小数取模：a - b*trunc(a/b)
	a, _ := number.ParseDec("5.5")
	m, _ = number.Mod(a, 2)
	if number.Format(m) != "1.5" {
		t.Errorf("5.5%%2 = %s", number.Format(m))
	}
	// 负数取模跟随被除数符号（trunc 语义）
	m, _ = number.Mod(-7, 3)
	if m != -1 {
		t.Errorf("-7%%3 = %v, want -1", m)
	}
}

func TestParseGeneric(t *testing.T) {
	if v, ok := number.Parse("42"); !ok || v != 42 {
		t.Errorf("Parse(42) = %v", v)
	}
	if v, ok := number.Parse(" 0.25 "); !ok || number.Format(v) != "0.25" {
		t.Errorf("Parse(0.25) = %v", v)
	}
	if _, ok := number.Parse("abc"); ok {
		t.Error("Parse(abc) 应失败")
	}
	if _, ok := number.Parse(""); ok {
		t.Error("Parse 空串应失败")
	}
}

func TestIsAndIsZero(t *testing.T) {
	if !number.Is(1) || !number.Is(number.Dec{Coef: 1, Scale: 1}) {
		t.Error("Is 判定错误")
	}
	if number.Is("1") {
		t.Error("字符串不是数字")
	}
	if !number.IsZero(0) {
		t.Error("0 应是零")
	}
	if number.IsZero(number.Dec{Coef: 1, Scale: 1}) {
		t.Error("0.1 不是零")
	}
}

func TestNonNumberOperands(t *testing.T) {
	if _, ok := number.Add("a", 1); ok {
		t.Error("字符串参与 Add 应失败")
	}
	if _, ok := number.Cmp(nil, 1); ok {
		t.Error("nil 参与 Cmp 应失败")
	}
}
