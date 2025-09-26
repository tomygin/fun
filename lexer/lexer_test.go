package lexer_test

import (
	"fun/lexer"
	"testing"
)

func TestLexer(t *testing.T) {
	lexer := lexer.NewLexer()

	code := `
	if x >= 10 {
		y := 3.14
		// this is a comment
		name = "hello world"
	}
	`

	tokens := lexer.Tokenize(code)

	for _, token := range tokens {
		t.Log(token)
	}
}
