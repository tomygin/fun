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
		panic(err)
	}

	tokens := lexer.NewLexer().Tokenize(string(source))
	ast := parser.NewParser().Parse(tokens)

	machine := vm.NewVM()
	// import 的相对路径以入口文件所在目录为基准
	machine.SetDir(filepath.Dir(os.Args[1]))

	if _, err := machine.Call(ast); err != nil {
		panic(err)
	}

}
