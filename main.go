package main

import (
	"fmt"
	"fun/lexer"
	"fun/parser"
	"fun/vm"
	"os"
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

	if _, err := vm.NewVM().Call(ast); err != nil {
		panic(err)
	}

}
