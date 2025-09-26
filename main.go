package main

import (
	"fun/lexer"
	"fun/parser"
	"fun/vm"
)

const code = `
a := 1
b := 2
c := 3

def double(n){
    return n * 2
    print('support return')
}

print(a * double(b) *c)

print('a'+'b')
`

func main() {
	tokens := lexer.NewLexer().Tokenize(code)
	ast := parser.NewParser().Parse(tokens)

	if _, err := vm.NewVM().Call(ast); err != nil {
		panic(err)
	}

}
