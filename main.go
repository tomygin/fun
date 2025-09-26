package main

import (
	"fun/lexer"
	"fun/parser"
	"fun/vm"
)

const code = `
a := 1

fun class(n){
	fun say(m){
		print(m)
	}
		fun init(){
			print(n)
		}
	
	return this
}

b := class(a)
b.say(2233)
b.init()

that := this

print(that.a)


`

func main() {
	tokens := lexer.NewLexer().Tokenize(code)
	ast := parser.NewParser().Parse(tokens)

	if _, err := vm.NewVM().Call(ast); err != nil {
		panic(err)
	}

}
