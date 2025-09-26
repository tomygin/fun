package main

import (
	"fmt"
	"fun/lexer"
	"fun/parser"
	"fun/vm"
)

const code = `
def fibonaci (n){
    if  (n < 2){
        1
    }else{
         (
            fibonaci(n - 1)
            +
            fibonaci((n - 2))
        )
    }
}


print(now())

times := 0

while (times < 15) {
    print(fibonaci(times))
    times++
}

print(now())

`

func main() {
	tokens := lexer.NewLexer().Tokenize(code)
	ast := parser.NewParser().Parse(tokens)

	resout, err := vm.NewVM().Call(ast)
	if err != nil {
		panic(err)
	}

	fmt.Println(resout)

}
