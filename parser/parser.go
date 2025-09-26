package parser

import (
	"fmt"
	"fun/lexer"
	"strconv"
)

type Parser struct {
	tokens []lexer.Token
	cursor int
}

func NewParser() *Parser {
	return &Parser{}
}

// Parse 主入口函数，
func (p *Parser) Parse(tokens []lexer.Token) []any {
	p.tokens = tokens
	p.cursor = 0

	result := []any{"begin"}
	parsed := p.parse()
	result = append(result, parsed...)
	return result
}

// current 获取当前token
func (p *Parser) current() lexer.Token {
	if p.cursor >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.cursor]
}

// back 获取后续token
func (p *Parser) back(step ...int) lexer.Token {
	s := 1
	if len(step) > 0 {
		s = step[0]
	}
	c := p.cursor + s
	if c >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[c]
}

// parse 解析所有token
func (p *Parser) parse() []any {
	var nodes []any
	for p.current().Type != lexer.EOF {
		nodes = append(nodes, p.work())
	}
	return nodes
}

// work 主要的解析逻辑
func (p *Parser) work() any {
	switch p.current().Type {
	case lexer.INT:
		return p.intLiteral()
	case lexer.FLOAT:
		return p.floatLiteral()
	case lexer.STRING:
		return p.stringLiteral()
	case lexer.IDENTIFIER:
		// 变量声明
		if p.back().Value == ":=" {
			return p.varDeclare()
		}
		// 变量修改
		if p.back().Value == "=" {
			return p.varAssign()
		}
		// 函数调用
		if p.back().Value == "(" {
			return p.varCall()
		}
		// 自增
		if p.back().Value == "++" {
			return p.selfAddSugar()
		}
		// 检查是否是表达式的一部分
		return p.parseExpression()
	case lexer.KEY:
		switch p.current().Value {
		case "while":
			return p.whileStatement()
		case "if":
			return p.ifStatement()
		case "def":
			return p.defStatement()
		case "return":
			return p.returnStatement()
		}
	case lexer.BRACKETS:
		if p.current().Value == "(" {
			return p.expressionSugar()
		}
	case lexer.OPERATOR:
		return p.operatorLiteral()
	}
	panic(fmt.Sprintf("unknown token %+v", p.current()))
}

// parseExpression 解析表达式（支持运算符优先级）
func (p *Parser) parseExpression() any {
	return p.parseComparison()
}

// parseComparison 解析比较运算符 (==, !=, <, >, <=, >=)
func (p *Parser) parseComparison() any {
	left := p.parseAddition()

	for p.cursor < len(p.tokens) && p.current().Type == lexer.OPERATOR {
		op := p.current().Value
		if op == "==" || op == "!=" || op == "<" || op == ">" || op == "<=" || op == ">=" {
			p.cursor++
			right := p.parseAddition()
			left = []any{p.convertOperator(op), left, right}
		} else {
			break
		}
	}

	return left
}

// parseAddition 解析加减运算符
func (p *Parser) parseAddition() any {
	left := p.parseMultiplication()

	for p.cursor < len(p.tokens) && p.current().Type == lexer.OPERATOR {
		op := p.current().Value
		if op == "+" || op == "-" {
			p.cursor++
			right := p.parseMultiplication()
			left = []any{p.convertOperator(op), left, right}
		} else {
			break
		}
	}

	return left
}

// parseMultiplication 解析乘除运算符
func (p *Parser) parseMultiplication() any {
	left := p.parsePrimary()

	for p.cursor < len(p.tokens) && p.current().Type == lexer.OPERATOR {
		op := p.current().Value
		if op == "*" || op == "/" {
			p.cursor++
			right := p.parsePrimary()
			left = []any{p.convertOperator(op), left, right}
		} else {
			break
		}
	}

	return left
}

// parsePrimary 解析基本表达式
func (p *Parser) parsePrimary() any {
	switch p.current().Type {
	case lexer.INT:
		return p.intLiteral()
	case lexer.FLOAT:
		return p.floatLiteral()
	case lexer.STRING:
		return p.stringLiteral()
	case lexer.IDENTIFIER:
		// 函数调用
		if p.back().Value == "(" {
			return p.varCall()
		}
		// 变量
		return p.varSelf()
	case lexer.BRACKETS:
		if p.current().Value == "(" {
			return p.expressionSugar()
		}
	}
	panic(fmt.Sprintf("unexpected token in expression: %+v", p.current()))
}

// convertOperator 转换操作符
func (p *Parser) convertOperator(op string) string {
	convertMap := map[string]string{
		"*":  "mul",
		"/":  "div",
		"-":  "sub",
		"+":  "add",
		">":  "gt",
		"<":  "lt",
		">=": "gte",
		"<=": "lte",
		"==": "eq",
		"!=": "neq",
	}

	if converted, ok := convertMap[op]; ok {
		return converted
	}
	return op
}

// intLiteral 解析整数
func (p *Parser) intLiteral() int {
	value, err := strconv.Atoi(p.current().Value)
	if err != nil {
		panic(fmt.Sprintf("invalid int: %s", p.current().Value))
	}
	p.cursor++
	return value
}

// floatLiteral 解析浮点数
func (p *Parser) floatLiteral() float64 {
	value, err := strconv.ParseFloat(p.current().Value, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid float: %s", p.current().Value))
	}
	p.cursor++
	return value
}

// stringLiteral 解析字符串
func (p *Parser) stringLiteral() string {
	value := p.current().Value
	p.cursor++
	return value
}

// operatorLiteral 解析操作符
func (p *Parser) operatorLiteral() string {
	value := p.current().Value

	convertMap := map[string]string{
		"*":  "mul",
		"/":  "div",
		"-":  "sub",
		"+":  "add",
		">":  "gt",
		"<":  "lt",
		">=": "gte",
		"<=": "lte",
		"==": "eq",
		"!=": "neq",
	}

	if converted, ok := convertMap[value]; ok {
		value = converted
	}

	p.cursor++
	return value
}

// varDeclare 变量声明
func (p *Parser) varDeclare() []any {
	name := p.current().Value
	p.cursor += 2
	value := p.parseExpression()
	return []any{"var", name, value}
}

// varAssign 变量赋值
func (p *Parser) varAssign() []any {
	name := p.current().Value
	p.cursor += 2
	value := p.parseExpression()
	return []any{"assign", name, value}
}

// varSelf 变量自身
func (p *Parser) varSelf() string {
	value := p.current().Value
	p.cursor++
	return value
}

// varCall 函数调用
func (p *Parser) varCall() []any {
	name := p.current().Value
	p.cursor += 2
	args := []any{name}

	for p.current().Value != ")" {
		args = append(args, p.parseExpression())
	}
	p.cursor++
	return args
}

// whileStatement while循环
func (p *Parser) whileStatement() []any {
	p.cursor++
	condition := p.parseExpression()

	if p.current().Value != "{" {
		panic("while loop must be '{'")
	}

	p.cursor++
	body := []any{"begin"}
	for p.current().Value != "}" {
		body = append(body, p.work())
	}
	p.cursor++
	return []any{"while", condition, body}
}

// ifStatement if语句
func (p *Parser) ifStatement() []any {
	p.cursor++
	condition := p.parseExpression()

	if p.current().Value != "{" {
		panic("if statement must be '{'")
	}

	p.cursor++
	body := []any{"begin"}
	for p.current().Value != "}" {
		body = append(body, p.work())
	}

	elseBody := []any{"begin"}
	p.cursor++

	if p.current().Value == "else" {
		p.cursor++
		if p.current().Value != "{" {
			panic("else statement must be '{'")
		}
		p.cursor++
		for p.current().Value != "}" {
			elseBody = append(elseBody, p.work())
		}
		p.cursor++
	}

	return []any{"if", condition, body, elseBody}
}

// defStatement 函数定义
func (p *Parser) defStatement() []any {
	p.cursor++

	if p.current().Type != lexer.IDENTIFIER {
		panic("def statement name must be identifier")
	}

	name := p.current().Value
	if p.back().Value != "(" {
		panic("def statement must be '('")
	}

	p.cursor += 2
	var args []any

	for p.current().Value != ")" {
		if p.current().Type != lexer.IDENTIFIER {
			panic("def statement arg must be identifier")
		}
		args = append(args, p.current().Value)
		p.cursor++
	}

	if p.back().Value != "{" {
		panic("def statement must be '{'")
	}

	p.cursor += 2
	body := []any{"begin"}
	for p.current().Value != "}" {
		body = append(body, p.work())
	}
	p.cursor++

	return []any{"def", name, args, body}
}

// returnStatement 处理 return 语句
func (p *Parser) returnStatement() []any {
	p.cursor++ // 跳过 return 关键字

	// 检查是否有返回值
	if p.current().Value == "}" || p.current().Type == lexer.EOF {
		// 没有返回值的 return
		return []any{"return", nil}
	}

	// 有返回值的 return
	value := p.parseExpression()
	return []any{"return", value}
}

// selfAddSugar 自增语法糖
func (p *Parser) selfAddSugar() []any {
	value := p.current().Value
	p.cursor += 2
	return []any{"assign", value, []any{"add", value, 1}}
}

// expressionSugar 表达式语法糖（括号表达式）
func (p *Parser) expressionSugar() any {
	p.cursor++ // 跳过 '('
	result := p.parseExpression()
	if p.current().Value != ")" {
		panic("expected ')' after expression")
	}
	p.cursor++ // 跳过 ')'
	return result
}
