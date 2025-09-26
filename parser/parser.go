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

// NewParser 创建新的parser实例
func NewParser() *Parser {
	return &Parser{}
}

// Parse 主入口函数，相当于Python的__call__
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
		// 表访问
		if p.back().Value == "[" {
			return p.tableSugar()
		}
		// 点访问
		if p.back().Value == "." {
			return p.tablePointSugar()
		}
		// 自增
		if p.back().Value == "++" {
			return p.selfAddSugar()
		}
		// 默认只有自己
		return p.varSelf()
	case lexer.KEY:
		switch p.current().Value {
		case "while":
			return p.whileStatement()
		case "if":
			return p.ifStatement()
		case "def":
			return p.defStatement()
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
	value := p.work()
	return []any{"var", name, value}
}

// varAssign 变量赋值
func (p *Parser) varAssign() []any {
	name := p.current().Value
	p.cursor += 2
	value := p.work()
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
		args = append(args, p.work())
	}
	p.cursor++
	return args
}

// whileStatement while循环
func (p *Parser) whileStatement() []any {
	p.cursor++
	condition := p.work()

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
	condition := p.work()

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

// tableSugar 表访问语法糖
func (p *Parser) tableSugar() []any {
	tableName := p.current().Value
	p.cursor += 2
	tableKey := p.work()

	if p.current().Value != "]" {
		panic("table sugar must be ']'")
	}

	// set table
	if p.back().Value == "=" {
		p.cursor += 2
		tableValue := p.work()
		return []any{"table_set", tableName, tableKey, tableValue}
	}

	// get table
	p.cursor++
	return []any{"table_get", tableName, tableKey}
}

// tablePointSugar 点访问语法糖
func (p *Parser) tablePointSugar() []any {
	tableName := p.current().Value
	p.cursor += 2

	var tableKey string
	if p.current().Type == lexer.IDENTIFIER {
		tableKey = "'" + p.current().Value + "'"
	} else {
		panic("table key must be identifier or number")
	}

	// set table
	if p.back().Value == "=" {
		p.cursor += 2
		tableValue := p.work()
		return []any{"table_set", tableName, tableKey, tableValue}
	}

	// get table
	p.cursor++
	return []any{"table_get", tableName, tableKey}
}

// selfAddSugar 自增语法糖
func (p *Parser) selfAddSugar() []any {
	value := p.current().Value
	p.cursor += 2
	return []any{"assign", value, []any{"add", value, 1}}
}

// expressionSugar 表达式语法糖
func (p *Parser) expressionSugar() any {
	p.cursor++
	var operations []string
	var elements []any
	isOperation := false

	// 操作符优先级顺序
	rank := []string{"mul", "div", "add", "sub", "gt", "lt", "gte", "lte", "eq", "neq"}

	for p.current().Value != ")" {
		ele := p.work()

		if isOperation {
			if strEle, ok := ele.(string); ok {
				operations = append(operations, strEle)
			}
		} else {
			elements = append(elements, ele)
		}

		isOperation = !isOperation
	}

	// 按照优先级构建
	for len(operations) > 0 {
		index := 0
		found := false

		// 找出优先级最高的操作符
		for _, op := range rank {
			for i, operation := range operations {
				if operation == op {
					index = i
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			panic("unknown operation, not in operations")
		}

		// 构建表达式树
		op := operations[index]
		operations = append(operations[:index], operations[index+1:]...)

		left := elements[index]
		right := elements[index+1]
		elements = append(elements[:index], elements[index+2:]...)
		elements = append(elements[:index], append([]any{[]any{op, left, right}}, elements[index:]...)...)
	}

	p.cursor++
	return elements[0]
}
