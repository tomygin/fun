package vm

import (
	"bytes"
	"encoding/json"
	"fun/number"
	"sort"
	"strconv"
	"strings"
)

// ----------------------------------------------------------------------------
// JSON：表 <-> JSON 文本
//
//   json.encode(v)  表/数组表/标量 -> JSON 字符串
//   json.decode(s)  JSON 字符串 -> 表/标量（数字走数字塔，0.1 依然精确）
//
// 约定：连续数字键 0..n-1 的表编码成 JSON 数组，其余编码成对象；
// 解码时 JSON 数组变成数字键表（配合 len/for 遍历）。
// ----------------------------------------------------------------------------

var jsonTable = map[string]any{
	// json.encode(v [, indent]) -> JSON 字符串；indent 为缩进空格数（美化输出）
	"encode": func(args ...any) any {
		if len(args) < 1 {
			return "null"
		}
		var b strings.Builder
		encodeJSON(&b, args[0])
		out := b.String()

		// 可选第二参数：缩进空格数
		if len(args) >= 2 {
			if n, ok := number.ToFloat(args[1]); ok && n > 0 {
				var pretty bytes.Buffer
				if err := json.Indent(&pretty, []byte(out), "", strings.Repeat(" ", int(n))); err == nil {
					return pretty.String()
				}
			}
		}
		return out
	},
	// json.decode(s) -> 值；非法 JSON 抛出错误（用 try 捕获），
	// 因此合法的 "null" 解码为 nil 与"解析失败"能明确区分。
	"decode": func(args ...any) (any, error) {
		if len(args) < 1 {
			return nil, &FunError{Value: "json.decode requires a string"}
		}
		s, ok := args[0].(string)
		if !ok {
			return nil, &FunError{Value: "json.decode requires a string"}
		}
		dec := json.NewDecoder(strings.NewReader(s))
		dec.UseNumber() // 数字保持字符串形态，交给数字塔精确解析
		var raw any
		if err := dec.Decode(&raw); err != nil {
			return nil, &FunError{Value: "invalid json: " + err.Error()}
		}
		return fromJSON(raw), nil
	},
}

// isArrayTable 判断表是否是"数组式"：键恰好为 "0".."n-1"
func isArrayTable(m map[string]any) bool {
	n := 0
	for k := range m {
		if len(k) > 0 && k[0] == '@' {
			continue
		}
		n++
	}
	if n == 0 {
		return false
	}
	for i := 0; i < n; i++ {
		if _, ok := m[strconv.Itoa(i)]; !ok {
			return false
		}
	}
	return true
}

// encodeJSON 把 Fun 值编码成 JSON 文本
func encodeJSON(b *strings.Builder, v any) {
	switch val := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if val {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case string:
		data, _ := json.Marshal(val)
		b.Write(data)
	case map[string]any:
		if isUserFunc(val) {
			b.WriteString("null") // 函数无法进 JSON
			return
		}
		if isArrayTable(val) {
			b.WriteByte('[')
			n := 0
			for k := range val {
				if len(k) > 0 && k[0] == '@' {
					continue
				}
				n++
			}
			for i := 0; i < n; i++ {
				if i > 0 {
					b.WriteByte(',')
				}
				encodeJSON(b, val[strconv.Itoa(i)])
			}
			b.WriteByte(']')
			return
		}
		keys := make([]string, 0, len(val))
		for k := range val {
			if len(k) > 0 && k[0] == '@' {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			kd, _ := json.Marshal(k)
			b.Write(kd)
			b.WriteByte(':')
			encodeJSON(b, val[k])
		}
		b.WriteByte('}')
	default:
		if number.Is(v) {
			b.WriteString(number.Format(v)) // 数字塔精确十进制输出
			return
		}
		data, _ := json.Marshal(formatValue(v))
		b.Write(data)
	}
}

// fromJSON 把 encoding/json 的产物转成 Fun 的值
func fromJSON(v any) any {
	switch val := v.(type) {
	case json.Number:
		if n, ok := number.Parse(string(val)); ok {
			return n // 走数字塔：0.1 解析成精确的 1/10
		}
		return string(val)
	case []any:
		out := make(map[string]any, len(val))
		for i, item := range val {
			out[strconv.Itoa(i)] = fromJSON(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, item := range val {
			out[k] = fromJSON(item)
		}
		return out
	}
	return v // string / bool / nil
}
