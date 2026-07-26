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

	tokens := lexer.NewLexer().Tokenize(string(source))
	ast := parser.NewParser().Parse(tokens) // ["begin", 语句...]

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
