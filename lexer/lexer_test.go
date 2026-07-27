package lexer_test

import (
	"fun/lexer"
	"testing"
)

// tokenize 返回去掉 EOF 的 (type, value) 序列
func tokenize(t *testing.T, code string) []lexer.Token {
	t.Helper()
	tokens := lexer.NewLexer().Tokenize(code)
	if len(tokens) == 0 || tokens[len(tokens)-1].Type != lexer.EOF {
		t.Fatalf("token 流必须以 EOF 结尾: %v", tokens)
	}
	return tokens[:len(tokens)-1]
}

// expectValues 断言 token 值序列
func expectValues(t *testing.T, code string, want ...string) {
	t.Helper()
	tokens := tokenize(t, code)
	if len(tokens) != len(want) {
		t.Fatalf("%q: token 数 %d, want %d: %v", code, len(tokens), len(want), tokens)
	}
	for i, w := range want {
		if tokens[i].Value != w {
			t.Errorf("%q: token[%d] = %q, want %q", code, i, tokens[i].Value, w)
		}
	}
}

func TestBasicTokens(t *testing.T) {
	expectValues(t, `a := 1 + 2.5`, "a", ":=", "1", "+", "2.5")
	expectValues(t, `x == y != z`, "x", "==", "y", "!=", "z")
	expectValues(t, `a >= b <= c`, "a", ">=", "b", "<=", "c")
}

func TestLogicalAndPipeOperators(t *testing.T) {
	// || 必须优先于 |，!= 必须优先于 !
	expectValues(t, `a || b | c && !d`, "a", "||", "b", "|", "c", "&&", "!", "d")
}

func TestAtIdentifier(t *testing.T) {
	// @ 前缀标识符（元编程空间）
	expectValues(t, `@add(1, 2)`, "@add", "(", "1", ",", "2", ")")
	tokens := tokenize(t, `@add`)
	if tokens[0].Type != lexer.IDENTIFIER {
		t.Errorf("@add 应是标识符, got %v", tokens[0].Type)
	}
}

func TestKeywords(t *testing.T) {
	tokens := tokenize(t, `if for else fun return this true false nil`)
	for _, tok := range tokens {
		if tok.Type != lexer.KEY {
			t.Errorf("%q 应是关键字, got %v", tok.Value, tok.Type)
		}
	}
	// 前缀撞关键字的标识符不能误判
	tokens = tokenize(t, `iffy forloop thistle`)
	for _, tok := range tokens {
		if tok.Type != lexer.IDENTIFIER {
			t.Errorf("%q 应是标识符, got %v", tok.Value, tok.Type)
		}
	}
}

func TestStringsWithEscapes(t *testing.T) {
	// 转义引号不能截断字符串
	tokens := tokenize(t, `"say \"hi\""`)
	if len(tokens) != 1 || tokens[0].Type != lexer.STRING {
		t.Fatalf("转义双引号解析错误: %v", tokens)
	}
	tokens = tokenize(t, `'it\'s'`)
	if len(tokens) != 1 || tokens[0].Type != lexer.STRING {
		t.Fatalf("转义单引号解析错误: %v", tokens)
	}
}

func TestTemplateString(t *testing.T) {
	tokens := tokenize(t, "`hello ${name}`")
	if len(tokens) != 1 || tokens[0].Type != lexer.STRING {
		t.Fatalf("模板字符串解析错误: %v", tokens)
	}
	// 跨多行
	tokens = tokenize(t, "`line1\nline2`")
	if len(tokens) != 1 {
		t.Fatalf("多行模板字符串解析错误: %v", tokens)
	}
}

func TestMultilineTemplateLineNumbers(t *testing.T) {
	// 多行模板字符串之后的 token 行号必须正确
	tokens := tokenize(t, "`a\nb\nc`\nx")
	last := tokens[len(tokens)-1]
	if last.Value != "x" || last.Line != 4 {
		t.Errorf("多行字符串后行号 = %d, want 4", last.Line)
	}
}

func TestCommentAtEOF(t *testing.T) {
	// 文件末尾无换行的注释不能 panic
	tokens := tokenize(t, "a := 1 // comment no newline")
	expectLen := 3
	if len(tokens) != expectLen {
		t.Errorf("EOF 注释: token 数 %d, want %d", len(tokens), expectLen)
	}
}

func TestNumbers(t *testing.T) {
	tokens := tokenize(t, `3.14`)
	if tokens[0].Type != lexer.FLOAT {
		t.Errorf("3.14 应是 FLOAT")
	}
	tokens = tokenize(t, `42`)
	if tokens[0].Type != lexer.INT {
		t.Errorf("42 应是 INT")
	}
}

func TestSemicolonAndBrackets(t *testing.T) {
	expectValues(t, `for i := 0; i < 3; i++ {}`,
		"for", "i", ":=", "0", ";", "i", "<", "3", ";", "i", "++", "{", "}")
}

func TestLineNumbers(t *testing.T) {
	tokens := tokenize(t, "a\nb\n\nc")
	wantLines := map[string]int{"a": 1, "b": 2, "c": 4}
	for _, tok := range tokens {
		if want, ok := wantLines[tok.Value]; ok && tok.Line != want {
			t.Errorf("%q 行号 = %d, want %d", tok.Value, tok.Line, want)
		}
	}
}
