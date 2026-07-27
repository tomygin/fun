package vm

import (
	"fmt"
	"fun/lexer"
	"fun/parser"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------------------------------------------------------
// 多文件编程：包的导入 / 导出
//
// 设计哲学：模块就是一张表 { }。
//   - import("xxx.fun") 会读取、解析并求值该文件，返回它的"导出表"。
//   - 导出有两种方式：
//       1. 隐式：模块顶层定义的所有名字（函数 / 变量）都会被打包成一张表。
//       2. 显式：模块顶层用 `return { ... }`（或 return this）自行决定导出什么。
//   - 相对路径以"当前正在执行的文件所在目录"为基准。
//   - 模块按绝对路径缓存，重复 import 只求值一次，也能容忍循环导入。
// ----------------------------------------------------------------------------

// interpolate 求值模板字符串里的 ${表达式} 插值。
// 实现是纯组合：把 ${} 里的文本重新走一遍 词法 -> 语法 -> 求值，
// 在当前环境里执行，所以能引用当前作用域的任何变量 / 调任何函数。
func (vm *VM) interpolate(s string) (string, error) {
	if !strings.Contains(s, "${") {
		return s, nil // 快速路径：没有插值
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] == '$' && i+1 < len(s) && s[i+1] == '{' {
			// 找到配对的 }（支持内层嵌套 { }，比如表字面量）
			depth := 1
			j := i + 2
			for j < len(s) && depth > 0 {
				switch s[j] {
				case '{':
					depth++
				case '}':
					depth--
				}
				j++
			}
			if depth != 0 {
				return "", fmt.Errorf("template string: unclosed ${")
			}
			exprSrc := s[i+2 : j-1]

			value, err := vm.evalSource(exprSrc)
			if err != nil {
				return "", err
			}
			b.WriteString(formatValue(value))
			i = j
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String(), nil
}

// parseModule 解析模块源码，panic 转错误
func parseModule(src string) (ast []any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("syntax error: %v", r)
		}
	}()
	tokens := lexer.NewLexer().Tokenize(src)
	return parser.NewParser().Parse(tokens), nil // ["begin", 语句...]
}

// evalSource 把一段源码走一遍 词法 -> 语法 -> 求值；
// 词法/语法的 panic 被转成普通错误（可被 try 捕获，也让顶层报错干净）。
func (vm *VM) evalSource(src string) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("syntax error: %v", r)
		}
	}()
	tokens := lexer.NewLexer().Tokenize(src)
	ast := parser.NewParser().Parse(tokens)
	return vm.Eval(ast)
}

// importModule 加载并求值一个模块文件，返回它导出的表。
func (vm *VM) importModule(path string) (any, error) {
	// 补全默认扩展名 .fun
	if filepath.Ext(path) == "" {
		path += ".fun"
	}

	// 相对路径以当前文件目录为基准
	full := path
	if !filepath.IsAbs(path) {
		full = filepath.Join(vm.dir, path)
	}
	full = filepath.Clean(full)

	// 命中缓存直接返回（同时避免循环导入无限递归）
	if cached, ok := vm.modules[full]; ok {
		return cached, nil
	}

	source, err := os.ReadFile(full)
	if err != nil {
		return nil, fmt.Errorf("import %q failed: %v", path, err)
	}

	// 模块的词法/语法错误转成普通错误（可被 try 捕获）
	ast, err := parseModule(string(source))
	if err != nil {
		return nil, fmt.Errorf("import %q: %v", path, err)
	}

	// 为模块创建独立环境，以内置函数环境为根
	moduleEnv := NewEnvironment(nil, vm.global)

	// 先在缓存里放一个空表占位，支持循环导入（拿到的是同一张表的引用）
	placeholder := make(map[string]any)
	vm.modules[full] = placeholder

	// 切换执行上下文到模块（保存并在结束后恢复）
	savedEnv, savedDir, savedReturning := vm.env, vm.dir, vm.returning
	vm.env = moduleEnv
	vm.dir = filepath.Dir(full)
	vm.returning = false

	// 逐条求值模块顶层语句；若顶层出现 return，则用其值作为显式导出
	var explicitExport any
	hasExplicit := false
	for _, stmt := range ast[1:] {
		result, err := vm.Eval(stmt)
		if err != nil {
			vm.env, vm.dir, vm.returning = savedEnv, savedDir, savedReturning
			delete(vm.modules, full)
			return nil, err
		}
		if vm.returning {
			explicitExport = result
			hasExplicit = true
			vm.returning = false
			break
		}
	}

	// 恢复调用方上下文
	vm.env, vm.dir, vm.returning = savedEnv, savedDir, savedReturning

	// 显式导出：直接用 return 的值
	if hasExplicit {
		vm.modules[full] = explicitExport
		return explicitExport, nil
	}

	// 隐式导出：把模块顶层的所有名字拷进占位表（跳过 @ 内部变量）
	for k, v := range moduleEnv.local {
		if strings.HasPrefix(k, "@") {
			continue
		}
		placeholder[k] = v
	}
	return placeholder, nil
}
