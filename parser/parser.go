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

// work 主要的解析逻辑：区分语句和表达式
func (p *Parser) work() any {
	switch p.current().Type {
	case lexer.IDENTIFIER:
		// 变量声明
		if p.back().Value == ":=" {
			return p.varDeclare()
		}
		// 自增
		if p.back().Value == "++" {
			return p.selfAddSugar()
		}
		// 赋值（支持 a = v / a.b = v / a[i] = v）
		if p.scanAssign() {
			return p.varAssign()
		}
		// 否则作为表达式求值
		return p.parseExpression()
	case lexer.KEY:
		switch p.current().Value {
		case "for":
			return p.forStatement()
		case "if":
			return p.ifStatement()
		case "fun":
			return p.defStatement()
		case "return":
			return p.returnStatement()
		}
		// this.x = v / this[k] = v 这类以 this 开头的属性赋值
		if p.current().Value == "this" && p.scanAssign() {
			return p.varAssign()
		}
		// this / true / false / nil 等作为表达式
		return p.parseExpression()
	default:
		// INT / FLOAT / STRING / ( / { / 一元运算符 等都是表达式
		return p.parseExpression()
	}
}

// parseExpression 解析表达式（支持运算符优先级）
// 优先级从低到高：or -> and -> 比较 -> 加减 -> 乘除模 -> 一元 -> 后缀 -> 基本
func (p *Parser) parseExpression() any {
	return p.parseLogicalOr()
}

// parseLogicalOr 解析逻辑或 (||)，短路由 VM 处理
func (p *Parser) parseLogicalOr() any {
	left := p.parseLogicalAnd()

	for p.cursor < len(p.tokens) && p.current().Type == lexer.OPERATOR && p.current().Value == "||" {
		p.cursor++
		right := p.parseLogicalAnd()
		left = []any{"or", left, right}
	}

	return left
}

// parseLogicalAnd 解析逻辑与 (&&)，短路由 VM 处理
func (p *Parser) parseLogicalAnd() any {
	left := p.parseComparison()

	for p.cursor < len(p.tokens) && p.current().Type == lexer.OPERATOR && p.current().Value == "&&" {
		p.cursor++
		right := p.parseComparison()
		left = []any{"and", left, right}
	}

	return left
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

// parseMultiplication 解析乘除模运算符
func (p *Parser) parseMultiplication() any {
	left := p.parseUnary()

	for p.cursor < len(p.tokens) && p.current().Type == lexer.OPERATOR {
		op := p.current().Value
		if op == "*" || op == "/" || op == "%" {
			p.cursor++
			right := p.parseUnary()
			left = []any{p.convertOperator(op), left, right}
		} else {
			break
		}
	}

	return left
}

// parseUnary 解析一元运算符 (-x, +x, not x)
func (p *Parser) parseUnary() any {
	cur := p.current()

	if cur.Type == lexer.OPERATOR && (cur.Value == "-" || cur.Value == "+") {
		p.cursor++
		operand := p.parseUnary()
		if cur.Value == "-" {
			// -x 等价于 0 - x
			return []any{"@sub", 0, operand}
		}
		return operand
	}

	if cur.Type == lexer.OPERATOR && cur.Value == "!" {
		p.cursor++
		return []any{"not", p.parseUnary()}
	}

	return p.parsePostfix()
}

// parsePostfix 解析后缀访问：属性 obj.property、方法 obj.method()、下标 obj[key]
func (p *Parser) parsePostfix() any {
	left := p.parsePrimary()

	for p.cursor < len(p.tokens) {
		cur := p.current()

		// 属性访问或方法调用
		if cur.Type == lexer.OPERATOR && cur.Value == "." {
			p.cursor++ // 跳过 '.'

			if p.current().Type != lexer.IDENTIFIER {
				panic("expected identifier after '.'")
			}

			property := p.current().Value
			p.cursor++

			// 检查是否是方法调用
			if p.cursor < len(p.tokens) && p.current().Value == "(" {
				p.cursor++ // 跳过 '('

				// 解析参数
				var args []any
				for p.current().Value != ")" {
					args = append(args, p.parseExpression())
					p.skipComma()
				}
				p.cursor++ // 跳过 ')'

				// 返回方法调用节点 ["method-call", object, method_name, ...args]
				methodCall := []any{"method-call", left, property}
				methodCall = append(methodCall, args...)
				left = methodCall
			} else {
				// 属性访问 ["property", object, property_name]
				left = []any{"property", left, property}
			}
			continue
		}

		// 下标访问 obj[key]
		if cur.Value == "[" {
			p.cursor++ // 跳过 '['
			index := p.parseExpression()
			if p.current().Value != "]" {
				panic("expected ']' after index")
			}
			p.cursor++ // 跳过 ']'
			left = []any{"index", left, index}
			continue
		}

		break
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
	case lexer.KEY:
		switch p.current().Value {
		case "this":
			result := p.current().Value
			p.cursor++
			return result
		case "true":
			p.cursor++
			return true
		case "false":
			p.cursor++
			return false
		case "nil":
			p.cursor++
			return []any{"nil"}
		case "fun":
			// 表达式位置的 fun 是匿名函数（可作为表的值 / 参数 / 返回值）
			return p.funExpr()
		}
	case lexer.BRACKETS:
		if p.current().Value == "(" {
			return p.expressionSugar()
		}
		if p.current().Value == "{" {
			return p.parseTable()
		}
	}
	panic(fmt.Sprintf("unexpected token in expression: %+v", p.current()))
}

// convertOperator 把符号运算符转成内置函数名。
// 名字统一加 @ 前缀，放进用户取不到的命名空间，
// 这样即使用户把函数命名为 add / sub 等也不会覆盖 + / - 运算。
func (p *Parser) convertOperator(op string) string {
	convertMap := map[string]string{
		"*":  "@mul",
		"/":  "@div",
		"%":  "@mod",
		"-":  "@sub",
		"+":  "@add",
		">":  "@gt",
		"<":  "@lt",
		">=": "@gte",
		"<=": "@lte",
		"==": "@eq",
		"!=": "@neq",
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

// varDeclare 变量声明
func (p *Parser) varDeclare() []any {
	name := p.current().Value
	p.cursor += 2
	value := p.parseExpression()
	return []any{"var", name, value}
}

// scanAssign 向前扫描，判断当前位置是否是赋值语句的左值
// 左值形如： (标识符 | this) (.ident | [expr])*  后面紧跟一个 '='
func (p *Parser) scanAssign() bool {
	i := p.cursor
	if i >= len(p.tokens) {
		return false
	}
	base := p.tokens[i]
	isBase := base.Type == lexer.IDENTIFIER ||
		(base.Type == lexer.KEY && base.Value == "this")
	if !isBase {
		return false
	}
	i++
	for i < len(p.tokens) {
		v := p.tokens[i].Value
		if v == "." {
			i += 2 // 跳过 '.' 和属性名
		} else if v == "[" {
			depth := 1
			i++
			for i < len(p.tokens) && depth > 0 {
				switch p.tokens[i].Value {
				case "[":
					depth++
				case "]":
					depth--
				}
				i++
			}
		} else {
			break
		}
	}
	return i < len(p.tokens) &&
		p.tokens[i].Type == lexer.OPERATOR &&
		p.tokens[i].Value == "="
}

// parseLValue 解析赋值左值，返回变量名(string) 或 ["property"/"index", ...] 节点
func (p *Parser) parseLValue() any {
	var left any = p.current().Value // 标识符名
	p.cursor++

	for {
		switch p.current().Value {
		case ".":
			p.cursor++
			prop := p.current().Value
			p.cursor++
			left = []any{"property", left, prop}
		case "[":
			p.cursor++
			index := p.parseExpression()
			if p.current().Value != "]" {
				panic("expected ']' in assignment target")
			}
			p.cursor++
			left = []any{"index", left, index}
		default:
			return left
		}
	}
}

// varAssign 变量赋值（支持 a = v / a.b = v / a[i] = v）
func (p *Parser) varAssign() []any {
	target := p.parseLValue()
	p.cursor++ // 跳过 '='
	value := p.parseExpression()
	return []any{"assign", target, value}
}

// parseTable 解析表字面量 { key: value  key2: value2 }（逗号被词法器忽略）
func (p *Parser) parseTable() []any {
	p.cursor++ // 跳过 '{'
	result := []any{"table"}

	for p.current().Value != "}" {
		// 键：标识符 或 字符串
		var key any
		switch p.current().Type {
		case lexer.IDENTIFIER:
			// 用引号包裹让 VM 把它当作字符串常量求值
			key = `"` + p.current().Value + `"`
			p.cursor++
		case lexer.STRING:
			key = p.stringLiteral()
		default:
			panic(fmt.Sprintf("table key must be identifier or string, got %+v", p.current()))
		}

		if p.current().Value != ":" {
			panic("expected ':' after table key")
		}
		p.cursor++ // 跳过 ':'

		value := p.parseExpression()
		result = append(result, key, value)
		p.skipComma()
	}

	p.cursor++ // 跳过 '}'
	return result
}

// varSelf 变量自身
func (p *Parser) varSelf() string {
	value := p.current().Value
	p.cursor++
	return value
}

// skipComma 跳过可选的逗号分隔符（逗号可有可无，保持宽松）
func (p *Parser) skipComma() {
	if p.current().Value == "," {
		p.cursor++
	}
}

// varCall 函数调用
func (p *Parser) varCall() []any {
	name := p.current().Value
	p.cursor += 2
	args := []any{name}

	for p.current().Value != ")" {
		args = append(args, p.parseExpression())
		p.skipComma()
	}
	p.cursor++
	return args
}

// parseBlock 解析 { ... } 语句块，返回 ["begin", ...]
func (p *Parser) parseBlock() []any {
	if p.current().Value != "{" {
		panic("expected '{' to start a block")
	}
	p.cursor++
	body := []any{"begin"}
	for p.current().Value != "}" {
		body = append(body, p.work())
	}
	p.cursor++
	return body
}

// hasSemicolonBeforeBlock 向前扫描，判断循环体 '{' 之前是否有分号，
// 用来区分三段式 for 和条件式 for
func (p *Parser) hasSemicolonBeforeBlock() bool {
	depth := 0
	for i := p.cursor; i < len(p.tokens); i++ {
		if p.tokens[i].Type == lexer.EOF {
			break
		}
		switch p.tokens[i].Value {
		case "(", "[":
			depth++
		case ")", "]":
			depth--
		case "{":
			if depth == 0 {
				return false // 到达循环体
			}
		case ";":
			if depth == 0 {
				return true
			}
		}
	}
	return false
}

// forStatement for 循环（借鉴 go：唯一的循环关键字，三种形式）
//
//	for { ... }                    无限循环
//	for cond { ... }               条件循环
//	for init; cond; post { ... }   三段式
func (p *Parser) forStatement() []any {
	p.cursor++ // 跳过 'for'

	// for { }  无限循环
	if p.current().Value == "{" {
		body := p.parseBlock()
		return []any{"while", true, body}
	}

	// for init; cond; post { }  三段式
	if p.hasSemicolonBeforeBlock() {
		init := p.work()
		if p.current().Value != ";" {
			panic("for: expected ';' after init clause")
		}
		p.cursor++

		condition := p.parseExpression()
		if p.current().Value != ";" {
			panic("for: expected ';' after condition")
		}
		p.cursor++

		post := p.work()
		body := p.parseBlock()

		// 脱糖为： { init; while cond { body...; post } }
		loopBody := make([]any, len(body))
		copy(loopBody, body)
		loopBody = append(loopBody, post)
		return []any{"begin", init, []any{"while", condition, loopBody}}
	}

	// for cond { }  条件循环
	condition := p.parseExpression()
	body := p.parseBlock()
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
		// else if：把后续的 if 语句作为 else 分支的唯一语句（链式条件）
		if p.current().Type == lexer.KEY && p.current().Value == "if" {
			elseBody = append(elseBody, p.ifStatement())
			return []any{"if", condition, body, elseBody}
		}
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

// parseParams 解析 ( a, b, c ) 形式的形参列表（逗号可选）
func (p *Parser) parseParams() []any {
	if p.current().Value != "(" {
		panic("fun statement must be '('")
	}
	p.cursor++ // 跳过 '('

	var args []any
	for p.current().Value != ")" {
		if p.current().Type != lexer.IDENTIFIER {
			panic("fun statement arg must be identifier")
		}
		args = append(args, p.current().Value)
		p.cursor++
		p.skipComma()
	}
	p.cursor++ // 跳过 ')'
	return args
}

// defStatement 具名函数定义： fun name(args) { body }
func (p *Parser) defStatement() []any {
	p.cursor++ // 跳过 'fun'

	if p.current().Type != lexer.IDENTIFIER {
		panic("fun statement name must be identifier")
	}
	name := p.current().Value
	p.cursor++ // 跳过函数名

	args := p.parseParams()
	body := p.parseBlock()

	return []any{"fun", name, args, body}
}

// funExpr 匿名函数表达式： fun(args) { body }
// 可以作为表的值、函数参数、返回值，从而让 { } 里能放"方法"
func (p *Parser) funExpr() []any {
	p.cursor++ // 跳过 'fun'
	args := p.parseParams()
	body := p.parseBlock()
	return []any{"fun-expr", args, body}
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
	return []any{"assign", value, []any{"@add", value, 1}}
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
