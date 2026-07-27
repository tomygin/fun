package main

import (
	"fmt"
	"fun/lexer"
	"fun/parser"
	"fun/vm"
	"os"
	"path/filepath"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("have fun")
		return
	}

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}

	// 词法/语法错误：干净地报错退出，不打 Go 堆栈
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "语法错误:", r)
			os.Exit(1)
		}
	}()

	tokens := lexer.NewLexer().Tokenize(string(source))
	ast := parser.NewParser().Parse(tokens)

	machine := vm.NewVM()
	// import 的相对路径以入口文件所在目录为基准
	machine.SetDir(filepath.Dir(os.Args[1]))

	if _, err := machine.Call(ast); err != nil {
		// 未被 try 捕获的错误一路冒泡到这里
		if fe, ok := err.(*vm.FunError); ok {
			fmt.Fprintln(os.Stderr, "未捕获的错误:", fe.Error())
		} else {
			fmt.Fprintln(os.Stderr, "运行时错误:", err)
		}
		os.Exit(1)
	}

}
