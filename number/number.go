// Package number 实现 Fun 的三层数字塔：
//
//	第 1 层 int（int64）      整数快路径，带溢出检测
//	第 2 层 Dec（定点小数）    十进制精确小数：值 = Coef / 10^Scale
//	第 3 层 *big.Rat（有理数） 兜底：溢出自动提升 / 无限循环小数
//
// 设计目标：0.1 + 0.2 == 0.3 精确成立；整数永不回绕、永不丢位；
// 常规整数与小数运算保持 int64 速度，只有真正需要时才落到 big.Rat。
//
// 精度声明（写给文档，也写给未来的自己）：
//   - 小数按十进制精确存储与运算，不存在二进制浮点误差。
//   - 定点快路径容纳约 18 位有效数字；超出自动切换到大有理数，仍然精确。
//   - 唯一的"近似"发生在无限循环小数的打印：四舍五入到小数点后 20 位；
//     内部值保持精确，后续运算不受打印影响。
package number

import (
	"math/big"
	"strconv"
	"strings"
)

// Dec 定点小数：值 = Coef / 10^Scale。
// 归一化约定：Scale >= 1 且 Coef 不以 0 结尾（0.50 归一为 (5,1)）；
// Scale == 0 的值会被降级成 int，因此 Dec 一定带小数部分。
type Dec struct {
	Coef  int64
	Scale int32
}

// maxScale 定点快路径允许的最大小数位数（对应 pow10 表）
const maxScale = 18

// ratPrintDigits 无限循环小数打印时保留的小数位数
const ratPrintDigits = 20

var pow10 = [maxScale + 1]int64{
	1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000,
	1000000000, 10000000000, 100000000000, 1000000000000, 10000000000000,
	100000000000000, 1000000000000000, 10000000000000000, 100000000000000000,
	1000000000000000000,
}

// ----------------------------------------------------------------------------
// 判定与转换
// ----------------------------------------------------------------------------

// Is 判断 v 是否是数字塔里的值（或可无损进入塔的宿主整数）
func Is(v any) bool {
	switch v.(type) {
	case int, int64, Dec, *big.Rat:
		return true
	}
	return false
}

// toInt64 把宿主整数统一成 int64
func toInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	}
	return 0, false
}

// ToRat 把塔内任意数字提升为有理数（新分配，不共享内部状态）
func ToRat(v any) (*big.Rat, bool) {
	if n, ok := toInt64(v); ok {
		return new(big.Rat).SetInt64(n), true
	}
	switch n := v.(type) {
	case Dec:
		return new(big.Rat).SetFrac64(n.Coef, pow10[n.Scale]), true
	case *big.Rat:
		return n, true
	}
	return nil, false
}

// ToFloat 近似成 float64（供字符串下标等宿主交互使用，不参与语言内运算）
func ToFloat(v any) (float64, bool) {
	if n, ok := toInt64(v); ok {
		return float64(n), true
	}
	switch n := v.(type) {
	case Dec:
		return float64(n.Coef) / float64(pow10[n.Scale]), true
	case *big.Rat:
		f, _ := n.Float64()
		return f, true
	}
	return 0, false
}

// IsZero 数字是否为零
func IsZero(v any) bool {
	if n, ok := toInt64(v); ok {
		return n == 0
	}
	switch n := v.(type) {
	case Dec:
		return n.Coef == 0
	case *big.Rat:
		return n.Sign() == 0
	}
	return false
}

// ----------------------------------------------------------------------------
// 字面量解析
// ----------------------------------------------------------------------------

// ParseInt 解析整数字面量；超出 int64 自动用大数（永不因溢出报错）
func ParseInt(s string) (any, error) {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return int(n), nil
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, strconv.ErrSyntax
	}
	return normalizeRat(r), nil
}

// ParseDec 解析小数字面量成精确十进制；位数过多自动用大有理数
func ParseDec(s string) (any, error) {
	t := strings.TrimPrefix(s, "+")
	dot := strings.IndexByte(t, '.')
	if dot >= 0 {
		digits := t[:dot] + t[dot+1:]
		scale := len(t) - dot - 1
		if scale <= maxScale {
			if coef, err := strconv.ParseInt(digits, 10, 64); err == nil {
				return normalizeDec(coef, int32(scale)), nil
			}
		}
	}
	// 快路径装不下：用大有理数精确表示
	r, ok := new(big.Rat).SetString(t)
	if !ok {
		return nil, strconv.ErrSyntax
	}
	return normalizeRat(r), nil
}

// Parse 解析任意数字字符串（整数或小数）
func Parse(s string) (any, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}
	if strings.ContainsAny(s, ".eE") {
		if v, err := ParseDec(s); err == nil {
			return v, true
		}
		// 科学计数法等交给 big.Rat
		if r, ok := new(big.Rat).SetString(s); ok {
			return normalizeRat(r), true
		}
		return nil, false
	}
	if v, err := ParseInt(s); err == nil {
		return v, true
	}
	return nil, false
}

// ----------------------------------------------------------------------------
// 归一化与降级（保持值永远停在能停的最快一层）
// ----------------------------------------------------------------------------

// normalizeDec 去掉尾零；小数位归零则降级成 int
func normalizeDec(coef int64, scale int32) any {
	for scale > 0 && coef%10 == 0 {
		coef /= 10
		scale--
	}
	if scale == 0 {
		return int(coef)
	}
	return Dec{Coef: coef, Scale: scale}
}

var bigOne = big.NewInt(1)
var bigTen = big.NewInt(10)

// normalizeRat 尝试把有理数降级：整数 -> int；有限小数 -> Dec；否则保持 Rat。
// 原理：分母只含 2 和 5 的因子时是有限十进制小数，
// 所需小数位数 = max(2 的个数, 5 的个数)。
func normalizeRat(r *big.Rat) any {
	if r.IsInt() {
		if r.Num().IsInt64() {
			return int(r.Num().Int64())
		}
		return r
	}
	d := new(big.Int).Set(r.Denom()) // Rat 的分母恒为正
	twos := 0
	for d.Bit(0) == 0 {
		d.Rsh(d, 1)
		twos++
	}
	fives := 0
	five := big.NewInt(5)
	q, m := new(big.Int), new(big.Int)
	for {
		q.QuoRem(d, five, m)
		if m.Sign() != 0 {
			break
		}
		d.Set(q)
		fives++
	}
	if d.Cmp(bigOne) != 0 {
		return r // 分母还有其他因子 => 无限循环小数，留在 Rat（内部精确）
	}
	scale := twos
	if fives > scale {
		scale = fives
	}
	if scale > maxScale {
		return r
	}
	// coef = num * 10^scale / denom，此处必然整除
	coef := new(big.Int).Mul(r.Num(), new(big.Int).Exp(bigTen, big.NewInt(int64(scale)), nil))
	coef.Quo(coef, r.Denom())
	if !coef.IsInt64() {
		return r
	}
	return normalizeDec(coef.Int64(), int32(scale))
}

// ----------------------------------------------------------------------------
// 溢出检测的 int64 原语
// ----------------------------------------------------------------------------

func addChk(a, b int64) (int64, bool) {
	s := a + b
	if (b > 0 && s < a) || (b < 0 && s > a) {
		return 0, false
	}
	return s, true
}

func subChk(a, b int64) (int64, bool) {
	s := a - b
	if (b < 0 && s < a) || (b > 0 && s > a) {
		return 0, false
	}
	return s, true
}

func mulChk(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	if c/b != a {
		return 0, false
	}
	return c, true
}

// alignDec 把两个 Dec 对齐到同一小数位；对齐溢出则失败（走 Rat）
func alignDec(a, b Dec) (int64, int64, int32, bool) {
	if a.Scale == b.Scale {
		return a.Coef, b.Coef, a.Scale, true
	}
	if a.Scale < b.Scale {
		diff := b.Scale - a.Scale
		if diff > maxScale {
			return 0, 0, 0, false
		}
		if c, ok := mulChk(a.Coef, pow10[diff]); ok {
			return c, b.Coef, b.Scale, true
		}
		return 0, 0, 0, false
	}
	diff := a.Scale - b.Scale
	if diff > maxScale {
		return 0, 0, 0, false
	}
	if c, ok := mulChk(b.Coef, pow10[diff]); ok {
		return a.Coef, c, a.Scale, true
	}
	return 0, 0, 0, false
}

// toDec 把 int / Dec 统一成 Dec 视图（Rat 不走这里）
func toDec(v any) (Dec, bool) {
	if n, ok := toInt64(v); ok {
		return Dec{Coef: n, Scale: 0}, true
	}
	if d, ok := v.(Dec); ok {
		return d, true
	}
	return Dec{}, false
}

// ----------------------------------------------------------------------------
// 四则运算 + 取模 + 比较（对外接口，输入输出都是塔内的 any）
// ----------------------------------------------------------------------------

// Add 精确加法
func Add(a, b any) (any, bool) {
	// int + int 快路径
	if x, ok := toInt64(a); ok {
		if y, ok := toInt64(b); ok {
			if s, ok := addChk(x, y); ok {
				return int(s), true
			}
			return ratOp(a, b, func(r, x, y *big.Rat) { r.Add(x, y) })
		}
	}
	// 定点快路径（int 与 Dec 混合也走这里）
	if x, ok := toDec(a); ok {
		if y, ok := toDec(b); ok {
			if xc, yc, s, ok := alignDec(x, y); ok {
				if sum, ok := addChk(xc, yc); ok {
					return normalizeDec(sum, s), true
				}
			}
			return ratOp(a, b, func(r, x, y *big.Rat) { r.Add(x, y) })
		}
	}
	return ratOp(a, b, func(r, x, y *big.Rat) { r.Add(x, y) })
}

// Sub 精确减法
func Sub(a, b any) (any, bool) {
	if x, ok := toInt64(a); ok {
		if y, ok := toInt64(b); ok {
			if s, ok := subChk(x, y); ok {
				return int(s), true
			}
			return ratOp(a, b, func(r, x, y *big.Rat) { r.Sub(x, y) })
		}
	}
	if x, ok := toDec(a); ok {
		if y, ok := toDec(b); ok {
			if xc, yc, s, ok := alignDec(x, y); ok {
				if diff, ok := subChk(xc, yc); ok {
					return normalizeDec(diff, s), true
				}
			}
			return ratOp(a, b, func(r, x, y *big.Rat) { r.Sub(x, y) })
		}
	}
	return ratOp(a, b, func(r, x, y *big.Rat) { r.Sub(x, y) })
}

// Mul 精确乘法
func Mul(a, b any) (any, bool) {
	if x, ok := toInt64(a); ok {
		if y, ok := toInt64(b); ok {
			if p, ok := mulChk(x, y); ok {
				return int(p), true
			}
			return ratOp(a, b, func(r, x, y *big.Rat) { r.Mul(x, y) })
		}
	}
	if x, ok := toDec(a); ok {
		if y, ok := toDec(b); ok {
			if x.Scale+y.Scale <= maxScale {
				if p, ok := mulChk(x.Coef, y.Coef); ok {
					return normalizeDec(p, x.Scale+y.Scale), true
				}
			}
			return ratOp(a, b, func(r, x, y *big.Rat) { r.Mul(x, y) })
		}
	}
	return ratOp(a, b, func(r, x, y *big.Rat) { r.Mul(x, y) })
}

// Div 精确除法：除得尽降回快层，除不尽保持精确有理数。
// 除数为零返回 (0, true) 仅作兜底 —— 语言层（@div）在此之前
// 已把除零转成可捕获的错误。
func Div(a, b any) (any, bool) {
	if !Is(a) || !Is(b) {
		return nil, false
	}
	if IsZero(b) {
		return 0, true
	}
	// int / int 整除快路径
	if x, ok := toInt64(a); ok {
		if y, ok := toInt64(b); ok && x%y == 0 {
			return int(x / y), true
		}
	}
	return ratOp(a, b, func(r, x, y *big.Rat) { r.Quo(x, y) })
}

// Mod 取模：a - b*trunc(a/b)。除数为零返回 (0, true) 仅作兜底 ——
// 语言层（@mod）在此之前已把模零转成可捕获的错误。
func Mod(a, b any) (any, bool) {
	if !Is(a) || !Is(b) {
		return nil, false
	}
	if IsZero(b) {
		return 0, true
	}
	if x, ok := toInt64(a); ok {
		if y, ok := toInt64(b); ok {
			return int(x % y), true // int64 取模不会溢出（除 MinInt64 % -1）
		}
	}
	xr, _ := ToRat(a)
	yr, _ := ToRat(b)
	q := new(big.Rat).Quo(xr, yr)
	trunc := new(big.Int).Quo(q.Num(), q.Denom())
	prod := new(big.Rat).Mul(yr, new(big.Rat).SetInt(trunc))
	return normalizeRat(new(big.Rat).Sub(xr, prod)), true
}

// ratOp 在有理数层执行运算并尝试降级
func ratOp(a, b any, op func(r, x, y *big.Rat)) (any, bool) {
	x, ok := ToRat(a)
	if !ok {
		return nil, false
	}
	y, ok := ToRat(b)
	if !ok {
		return nil, false
	}
	r := new(big.Rat)
	op(r, x, y)
	return normalizeRat(r), true
}

// Cmp 精确比较：返回 -1 / 0 / 1
func Cmp(a, b any) (int, bool) {
	// int 快路径
	if x, ok := toInt64(a); ok {
		if y, ok := toInt64(b); ok {
			switch {
			case x < y:
				return -1, true
			case x > y:
				return 1, true
			}
			return 0, true
		}
	}
	// Dec 快路径
	if x, ok := toDec(a); ok {
		if y, ok := toDec(b); ok {
			if xc, yc, _, ok := alignDec(x, y); ok {
				switch {
				case xc < yc:
					return -1, true
				case xc > yc:
					return 1, true
				}
				return 0, true
			}
		}
	}
	xr, ok := ToRat(a)
	if !ok {
		return 0, false
	}
	yr, ok := ToRat(b)
	if !ok {
		return 0, false
	}
	return xr.Cmp(yr), true
}

// ----------------------------------------------------------------------------
// 打印
// ----------------------------------------------------------------------------

// Format 把塔内数字转成字符串。
// int / Dec / 有限小数：精确输出；
// 无限循环小数：四舍五入到小数点后 ratPrintDigits 位（内部值仍精确）。
func Format(v any) string {
	if n, ok := toInt64(v); ok {
		return strconv.FormatInt(n, 10)
	}
	switch n := v.(type) {
	case Dec:
		neg := n.Coef < 0
		c := n.Coef
		if neg {
			c = -c
		}
		s := strconv.FormatInt(c, 10)
		for int32(len(s)) <= n.Scale {
			s = "0" + s
		}
		cut := int32(len(s)) - n.Scale
		out := s[:cut] + "." + s[cut:]
		if neg {
			out = "-" + out
		}
		return out
	case *big.Rat:
		if n.IsInt() {
			return n.Num().String()
		}
		out := n.FloatString(ratPrintDigits)
		out = strings.TrimRight(out, "0")
		out = strings.TrimSuffix(out, ".")
		return out
	}
	return ""
}
