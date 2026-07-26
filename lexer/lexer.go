package lexer

import (
	"fmt"
	"regexp"
	"strings"
)

type TokenType int

const (
	KEY TokenType = iota + 1
	IDENTIFIER
	OPERATOR
	BRACKETS
	FLOAT
	INT
	STRING
	BLANK
	EOF
)

// String 方法让TokenType可以打印
func (t TokenType) String() string {
	switch t {
	case KEY:
		return "key"
	case IDENTIFIER:
		return "identifier"
	case OPERATOR:
		return "operator"
	case BRACKETS:
		return "brackets"
	case FLOAT:
		return "float"
	case INT:
		return "int"
	case STRING:
		return "string"
	case BLANK:
		return "blank"
	case EOF:
		return "eof"
	default:
		return "unknown"
	}
}

type Token struct {
	Type  TokenType
	Value string
	Line  int
}

// String 方法让Token可以打印
func (t Token) String() string {
	return fmt.Sprintf("type[%s] --> value[%s] in line %d", t.Type, t.Value, t.Line)
}

// TokenPattern 定义正则表达式和对应的token类型
type TokenPattern struct {
	Pattern string
	Type    TokenType
	Regex   *regexp.Regexp
}

// Lexer 词法分析器
type Lexer struct {
	code     string
	cursor   int
	tokens   []Token
	line     int
	patterns []TokenPattern
}

func NewLexer() *Lexer {
	lexer := &Lexer{
		line: 1,
	}

	// 定义token匹配规则（注意顺序很重要）
	patterns := []TokenPattern{
		// 注释（必须在除法之前，行尾或文件末尾都可以）
		{`^//[^\n]*`, BLANK, nil},
		// 括号
		{`^[\(\)\{\}\[\]]`, BRACKETS, nil},
		// 浮点数（必须在整数之前）
		{`^[+-]?\d+\.\d+`, FLOAT, nil},
		// 整数
		{`^[+-]?\d+`, INT, nil},
		// 空白字符
		{`^\s+`, BLANK, nil},
		// 逗号（参数/元素分隔符，作为独立 token 以消除歧义）
		{`^,`, OPERATOR, nil},
		// 分号（三段式 for 的子句分隔符）
		{`^;`, OPERATOR, nil},
		// 字符串（单引号）
		{`^'.*?'`, STRING, nil},
		// 字符串（双引号）
		{`^".*?"`, STRING, nil},
		// 比较与逻辑操作符（必须在单个字符操作符之前，!= 要在 ! 之前）
		{`^(>=|<=|!=|==|&&|\|\|)`, OPERATOR, nil},
		// 赋值操作符（必须在冒号之前）
		{`^:?=`, OPERATOR, nil},
		// 冒号（表字面量的键值分隔符）
		{`^:`, OPERATOR, nil},
		// 算术操作符（含取模）
		{`^[\+\-\*\/%]+`, OPERATOR, nil},
		// 其他操作符（含逻辑非 !）
		{`^[<>\.!]`, OPERATOR, nil},
		// 关键字（必须在标识符之前）
		{`^(if|for|else|fun|return|this|true|false|nil)\b`, KEY, nil},
		// 标识符
		{`^[a-zA-Z_][a-zA-Z0-9_]*`, IDENTIFIER, nil},
	}

	// 编译正则表达式
	for i, pattern := range patterns {
		regex, err := regexp.Compile(pattern.Pattern)
		if err != nil {
			panic(fmt.Sprintf("Invalid regex pattern: %s, error: %v", pattern.Pattern, err))
		}
		patterns[i].Regex = regex
	}

	lexer.patterns = patterns
	return lexer
}

// IsEOF 检查是否到达文件末尾
func (l *Lexer) IsEOF() bool {
	return l.cursor >= len(l.code)
}

// Tokenize 对代码进行词法分析
func (l *Lexer) Tokenize(code string) []Token {
	// 初始化
	l.code = code
	l.cursor = 0
	l.tokens = []Token{}
	l.line = 1

	for !l.IsEOF() {
		token := l.lex()
		if token.Type != BLANK {
			l.tokens = append(l.tokens, token)
		}
	}

	// 添加EOF token
	l.tokens = append(l.tokens, Token{
		Type:  EOF,
		Value: "eof",
		Line:  l.line,
	})

	return l.tokens
}

// lex 分析单个token
func (l *Lexer) lex() Token {
	if l.IsEOF() {
		return Token{Type: EOF, Value: "eof", Line: l.line}
	}

	remainingCode := l.code[l.cursor:]

	// 尝试匹配每个模式
	for _, pattern := range l.patterns {
		match := pattern.Regex.FindString(remainingCode)
		if match != "" {
			// 更新游标
			l.cursor += len(match)

			// 如果是空白字符，更新行号
			if pattern.Type == BLANK {
				l.line += strings.Count(match, "\n")
			}

			return Token{
				Type:  pattern.Type,
				Value: match,
				Line:  l.line,
			}
		}
	}

	// 如果没有匹配到任何模式，抛出错误
	panic(fmt.Sprintf("line %d with error\t\t\ncurrent >'%s'", l.line, remainingCode))
}
