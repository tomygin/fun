package vm

import (
	"bufio"
	"fmt"
	"fun/number"
	"math"
	"math/big"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// isUserFunc 判断一个表是否其实是用户定义的函数
func isUserFunc(m map[string]any) bool {
	_, hasParams := m["params"]
	_, hasBody := m["body"]
	return hasParams && hasBody
}

// formatValue 把 Fn 的值转成可读字符串，供 print/str 使用
func formatValue(value any) string {
	switch v := value.(type) {
	case nil:
		return "nil"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return v
	case *Coroutine:
		return "<coroutine>"
	case number.Dec, *big.Rat:
		// 数字塔：定点小数 / 大数，精确十进制输出
		return number.Format(v)
	case float64:
		// 宿主传入的浮点数（语言内运算不再产生 float64）
		if v == math.Trunc(v) && !math.IsInf(v, 0) {
			return strconv.FormatFloat(v, 'f', -1, 64)
		}
		return strconv.FormatFloat(v, 'g', -1, 64)
	case map[string]any:
		if isUserFunc(v) {
			return "<function>"
		}
		// 跳过 @ 开头的内部字段，键有序输出
		keys := make([]string, 0, len(v))
		for k := range v {
			if len(k) > 0 && k[0] == '@' {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		b.WriteString("{")
		for i, k := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(k)
			b.WriteString(": ")
			if child, ok := v[k].(map[string]any); ok {
				// 嵌套表：函数显示为 <function>，其它只显示 {...} 避免深层/循环递归
				if isUserFunc(child) {
					b.WriteString("<function>")
				} else {
					b.WriteString("{...}")
				}
			} else {
				b.WriteString(formatValue(v[k]))
			}
		}
		b.WriteString("}")
		return b.String()
	default:
		return fmt.Sprint(v)
	}
}

// valuesEqual 判断相等：数字走数字塔精确比较（int / 定点小数 / 大数
// 跨层可相等，0.1+0.2 == 0.3 为真），其它类型用深度比较。
func valuesEqual(a, b any) bool {
	if c, ok := number.Cmp(a, b); ok {
		return c == 0
	}
	if isNumber(a) && isNumber(b) {
		// 宿主 float 等塔外数字的兜底比较
		na, _ := toNumber(a)
		nb, _ := toNumber(b)
		return na == nb
	}
	return reflect.DeepEqual(a, b)
}

// typeName 返回值的类型名
func typeName(value any) string {
	switch value.(type) {
	case nil:
		return "nil"
	case bool:
		return "bool"
	case string:
		return "string"
	case *Coroutine:
		return "coroutine"
	case map[string]any:
		if isUserFunc(value.(map[string]any)) {
			return "function"
		}
		return "table"
	default:
		if isNumber(value) {
			return "number"
		}
		if reflect.TypeOf(value) != nil && reflect.TypeOf(value).Kind() == reflect.Func {
			return "function"
		}
		return "unknown"
	}
}

// ----------------------------------------------------------------------------
// Interface (built-in functions)
// ----------------------------------------------------------------------------

var interfaceFunctions = map[string]any{
	"VERSION": "1.0.0",
	"@neq": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		return !valuesEqual(args[0], args[1])
	},
	"@eq": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		return valuesEqual(args[0], args[1])
	},
	"@gt": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		if c, ok := number.Cmp(args[0], args[1]); ok {
			return c > 0
		}
		return false
	},
	"@gte": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		if c, ok := number.Cmp(args[0], args[1]); ok {
			return c >= 0
		}
		return false
	},
	"@lte": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		if c, ok := number.Cmp(args[0], args[1]); ok {
			return c <= 0
		}
		return false
	},
	"@lt": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		if c, ok := number.Cmp(args[0], args[1]); ok {
			return c < 0
		}
		return false
	},
	"@add": func(args ...any) any {
		if len(args) < 2 {
			return 0
		}
		// 任一操作数是字符串则做字符串拼接（借鉴 python/js 的 +）
		_, isStr1 := args[0].(string)
		_, isStr2 := args[1].(string)
		if isStr1 || isStr2 {
			return formatValue(args[0]) + formatValue(args[1])
		}
		if r, ok := number.Add(args[0], args[1]); ok {
			return r
		}
		return 0
	},
	"@sub": func(args ...any) any {
		if len(args) < 2 {
			return 0
		}
		if r, ok := number.Sub(args[0], args[1]); ok {
			return r
		}
		return 0
	},
	"@mul": func(args ...any) any {
		if len(args) < 2 {
			return 1
		}
		if r, ok := number.Mul(args[0], args[1]); ok {
			return r
		}
		return 1
	},
	"@div": func(args ...any) any {
		if len(args) < 2 {
			return 0
		}
		if r, ok := number.Div(args[0], args[1]); ok {
			return r
		}
		return 0
	},
	"@mod": func(args ...any) any {
		if len(args) < 2 {
			return 0
		}
		if r, ok := number.Mod(args[0], args[1]); ok {
			return r
		}
		return 0
	},
	"print": func(args ...any) any {
		parts := make([]string, 0, len(args))
		for _, arg := range args {
			parts = append(parts, formatValue(arg))
		}
		fmt.Println(strings.Join(parts, ""))
		return nil
	},
	"now": func(args ...any) any {
		return time.Now().String()
	},
	// len 返回字符串长度或表的键数量
	"len": func(args ...any) any {
		if len(args) < 1 {
			return 0
		}
		switch v := args[0].(type) {
		case string:
			return len([]rune(v))
		case map[string]any:
			n := 0
			for k := range v {
				if len(k) > 0 && k[0] == '@' {
					continue
				}
				n++
			}
			return n
		}
		return 0
	},
	// type 返回值的类型名："number" "string" "bool" "table" "function" "nil"
	"type": func(args ...any) any {
		if len(args) < 1 {
			return "nil"
		}
		return typeName(args[0])
	},
	// str 把任意值转成字符串
	"str": func(args ...any) any {
		if len(args) < 1 {
			return ""
		}
		return formatValue(args[0])
	},
	// num 把值转成数字（失败返回 nil）
	"num": func(args ...any) any {
		if len(args) < 1 {
			return nil
		}
		// 已经是塔内数字：原样返回
		if number.Is(args[0]) {
			return args[0]
		}
		// 字符串按精确十进制解析（"0.1" 就是精确的 1/10）
		if s, ok := args[0].(string); ok {
			if v, ok := number.Parse(s); ok {
				return v
			}
			return nil
		}
		if n, ok := toNumber(args[0]); ok {
			return n
		}
		return nil
	},
	// bool 把值转成布尔
	"bool": func(args ...any) any {
		if len(args) < 1 {
			return false
		}
		return toBool(args[0])
	},
	// has 判断表是否包含某个键
	"has": func(args ...any) any {
		if len(args) < 2 {
			return false
		}
		if m, ok := args[0].(map[string]any); ok {
			_, exists := m[formatValue(args[1])]
			return exists
		}
		return false
	},
	// del 从表中删除某个键，返回该表
	"del": func(args ...any) any {
		if len(args) < 2 {
			return nil
		}
		if m, ok := args[0].(map[string]any); ok {
			delete(m, formatValue(args[1]))
			return m
		}
		return args[0]
	},
	// clone 浅拷贝一张表 —— 面向对象的"实例化"：从原型表复制出实例
	"clone": func(args ...any) any {
		if len(args) < 1 {
			return nil
		}
		if m, ok := args[0].(map[string]any); ok {
			out := make(map[string]any, len(m))
			for k, v := range m {
				out[k] = v
			}
			return out
		}
		return args[0]
	},
	// merge 把后面各表的键依次覆盖进第一张表，返回第一张表。
	// 优先级：越靠后的表优先级越高 —— merge(a, b, c) 中
	// c 覆盖 b，b 覆盖 a，同名键以最后出现的为准。
	// 面向对象的"继承 / mixin"：merge(clone(父类), { 子类字段... })
	"merge": func(args ...any) any {
		if len(args) < 1 {
			return nil
		}
		dst, ok := args[0].(map[string]any)
		if !ok {
			return args[0]
		}
		for _, src := range args[1:] {
			if m, ok := src.(map[string]any); ok {
				for k, v := range m {
					dst[k] = v
				}
			}
		}
		return dst
	},
	// keys 返回表所有键组成的新表（键为 0..n-1），配合 len 可遍历
	"keys": func(args ...any) any {
		out := make(map[string]any)
		if len(args) < 1 {
			return out
		}
		if m, ok := args[0].(map[string]any); ok {
			names := make([]string, 0, len(m))
			for k := range m {
				if len(k) > 0 && k[0] == '@' {
					continue
				}
				names = append(names, k)
			}
			sort.Strings(names)
			for i, k := range names {
				out[strconv.Itoa(i)] = k
			}
		}
		return out
	},
	// upper/lower 字符串大小写转换
	"upper": func(args ...any) any {
		if len(args) < 1 {
			return ""
		}
		return strings.ToUpper(formatValue(args[0]))
	},
	"lower": func(args ...any) any {
		if len(args) < 1 {
			return ""
		}
		return strings.ToLower(formatValue(args[0]))
	},
	// assert 断言，失败则 panic
	"assert": func(args ...any) any {
		if len(args) < 1 {
			return nil
		}
		if !toBool(args[0]) {
			msg := "assertion failed"
			if len(args) >= 2 {
				msg = formatValue(args[1])
			}
			panic(msg)
		}
		return true
	},
	// costatus 返回协程状态："suspended" / "running" / "dead"
	"costatus": func(args ...any) any {
		if len(args) < 1 {
			return "nil"
		}
		if co, ok := args[0].(*Coroutine); ok {
			return co.status
		}
		return "nil"
	},
	// input 从标准输入读取一行（可带提示语）
	"input": func(args ...any) any {
		if len(args) >= 1 {
			fmt.Print(formatValue(args[0]))
		}
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		return strings.TrimRight(line, "\r\n")
	},
}
