package parser

import (
	"baptisia/ast"
	"baptisia/lexer"
	"fmt"
)

type Parser struct {
	l       *lexer.Lexer
	current lexer.Token
	peek    lexer.Token
	Errors  []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.current = p.l.NextToken()
	p.peek = p.l.NextToken()
	return p
}

func (p *Parser) advance() {
	p.current = p.peek
	p.peek = p.l.NextToken()
}

func (p *Parser) expect(t lexer.TokenType) {
	if p.current.Type != t {
		msg := fmt.Sprintf("line %d: expected %s but got %s (%q)",
			p.current.Line, t, p.current.Type, p.current.Literal)
		p.Errors = append(p.Errors, msg)
	}
	p.advance()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Device = p.parseDevice()
	return program
}

func (p *Parser) parseConst() *ast.ConstNode {
	node := &ast.ConstNode{Line: p.current.Line}
	p.expect(lexer.TOKEN_CONST)
	p.expect(lexer.TOKEN_LBRACE)

	for p.current.Type != lexer.TOKEN_RBRACE && p.current.Type != lexer.TOKEN_EOF {
		decl := ast.ConstDecl{Line: p.current.Line}
		decl.Name = p.current.Literal
		p.advance()
		p.expect(lexer.TOKEN_ASSIGN)
		decl.Value = p.current.Literal
		p.advance()
		node.Constants = append(node.Constants, decl)
	}

	p.expect(lexer.TOKEN_RBRACE)
	return node
}

func (p *Parser) parseStates() *ast.StatesNode {
	node := &ast.StatesNode{Line: p.current.Line}
	p.expect(lexer.TOKEN_STATES)
	p.expect(lexer.TOKEN_LBRACE)

	for p.current.Type != lexer.TOKEN_RBRACE && p.current.Type != lexer.TOKEN_EOF {
		node.Names = append(node.Names, p.current.Literal)
		p.advance()
	}

	p.expect(lexer.TOKEN_RBRACE)
	return node
}

func (p *Parser) parseDevice() *ast.DeviceNode {
	node := &ast.DeviceNode{Line: p.current.Line}

	p.expect(lexer.TOKEN_DEVICE)
	node.Name = p.current.Literal
	p.advance()
	p.expect(lexer.TOKEN_COLON)
	node.Target = p.current.Literal
	p.advance()
	p.expect(lexer.TOKEN_LBRACE)

	for p.current.Type != lexer.TOKEN_RBRACE && p.current.Type != lexer.TOKEN_EOF {
		switch p.current.Type {
		case lexer.TOKEN_WATCHDOG:
			p.advance()
			p.expect(lexer.TOKEN_COLON)
			node.Watchdog = p.current.Literal
			p.advance()
		case lexer.TOKEN_CYCLE:
			p.advance()
			p.expect(lexer.TOKEN_COLON)
			node.Cycle = p.current.Literal
			p.advance()
		case lexer.TOKEN_VARS:
			node.Vars = p.parseVars()
		case lexer.TOKEN_BOOT:
			node.Boot = p.parseBoot()
		case lexer.TOKEN_INPUTS:
			node.Inputs = p.parseInputs()
		case lexer.TOKEN_OUTPUTS:
			node.Outputs = p.parseOutputs()
		case lexer.TOKEN_SAFETY:
			node.Safety = p.parseSafety()
		case lexer.TOKEN_FAILSAFE:
			node.Failsafe = p.parseFailsafe()
		case lexer.TOKEN_CONTROL:
			node.Control = p.parseControl()
		case lexer.TOKEN_CONST:
			node.Consts = p.parseConst()
		case lexer.TOKEN_STATES:
			node.States = p.parseStates()
		default:
			p.Errors = append(p.Errors,
				fmt.Sprintf("line %d: unknown block or keyword %q",
					p.current.Line, p.current.Literal))
			p.advance()
		}
	}

	p.expect(lexer.TOKEN_RBRACE)
	return node
}

func (p *Parser) parseVars() *ast.VarsNode {
	node := &ast.VarsNode{Line: p.current.Line}
	p.expect(lexer.TOKEN_VARS)
	p.expect(lexer.TOKEN_LBRACE)

	for p.current.Type != lexer.TOKEN_RBRACE && p.current.Type != lexer.TOKEN_EOF {
		decl := ast.VarDecl{Line: p.current.Line}

		if p.current.Type == lexer.TOKEN_VOL {
			decl.Volatile = true
			p.advance()
		}

		decl.TypeName = string(p.current.Type)
		p.advance()
		decl.Name = p.current.Literal
		p.advance()
		p.expect(lexer.TOKEN_ASSIGN)
		decl.Value = p.current.Literal
		p.advance()

		node.Vars = append(node.Vars, decl)
	}

	p.expect(lexer.TOKEN_RBRACE)
	return node
}

func (p *Parser) parseBoot() *ast.BootNode {
	node := &ast.BootNode{Line: p.current.Line}
	p.expect(lexer.TOKEN_BOOT)
	p.expect(lexer.TOKEN_LBRACE)

	for p.current.Type != lexer.TOKEN_RBRACE && p.current.Type != lexer.TOKEN_EOF {
		if stmt := p.parseStatement(); stmt != nil {
			node.Statements = append(node.Statements, stmt)
		}
	}

	p.expect(lexer.TOKEN_RBRACE)
	return node
}

func (p *Parser) parseInputs() *ast.InputsNode {
	node := &ast.InputsNode{Line: p.current.Line}
	p.expect(lexer.TOKEN_INPUTS)
	p.expect(lexer.TOKEN_LBRACE)

	for p.current.Type != lexer.TOKEN_RBRACE && p.current.Type != lexer.TOKEN_EOF {
		sa := ast.SensorAssign{Line: p.current.Line}
		sa.Name = p.current.Literal
		p.advance()
		p.expect(lexer.TOKEN_ASSIGN)
		p.expect(lexer.TOKEN_VAL)
		p.expect(lexer.TOKEN_LPAREN)
		p.expect(lexer.TOKEN_SENSOR)
		p.expect(lexer.TOKEN_LPAREN)
		sa.Pin = p.current.Literal
		p.advance()
		p.expect(lexer.TOKEN_RPAREN)
		p.expect(lexer.TOKEN_COMMA)
		p.expect(lexer.TOKEN_IDENT)
		p.expect(lexer.TOKEN_COLON)
		sa.Min = p.current.Literal
		p.advance()
		p.expect(lexer.TOKEN_COMMA)
		p.expect(lexer.TOKEN_IDENT)
		p.expect(lexer.TOKEN_COLON)
		sa.Max = p.current.Literal
		p.advance()
		p.expect(lexer.TOKEN_RPAREN)

		node.Inputs = append(node.Inputs, sa)
	}

	p.expect(lexer.TOKEN_RBRACE)
	return node
}

func (p *Parser) parseOutputs() *ast.OutputsNode {
	node := &ast.OutputsNode{Line: p.current.Line}
	p.expect(lexer.TOKEN_OUTPUTS)
	p.expect(lexer.TOKEN_LBRACE)

	for p.current.Type != lexer.TOKEN_RBRACE && p.current.Type != lexer.TOKEN_EOF {
		ad := ast.ActuatorDecl{Line: p.current.Line}
		ad.Name = p.current.Literal
		p.advance()
		p.expect(lexer.TOKEN_ASSIGN)
		p.expect(lexer.TOKEN_ACTUATOR)
		p.expect(lexer.TOKEN_LPAREN)
		ad.Pin = p.current.Literal
		p.advance()
		p.expect(lexer.TOKEN_RPAREN)

		node.Outputs = append(node.Outputs, ad)
	}

	p.expect(lexer.TOKEN_RBRACE)
	return node
}

func (p *Parser) parseSafety() *ast.SafetyNode {
	node := &ast.SafetyNode{Line: p.current.Line}
	p.expect(lexer.TOKEN_SAFETY)
	p.expect(lexer.TOKEN_LBRACE)

	for p.current.Type != lexer.TOKEN_RBRACE && p.current.Type != lexer.TOKEN_EOF {
		if stmt := p.parseStatement(); stmt != nil {
			node.Statements = append(node.Statements, stmt)
		}
	}

	p.expect(lexer.TOKEN_RBRACE)
	return node
}

func (p *Parser) parseFailsafe() *ast.FailsafeNode {
	node := &ast.FailsafeNode{Line: p.current.Line}
	p.expect(lexer.TOKEN_FAILSAFE)
	p.expect(lexer.TOKEN_LBRACE)

	for p.current.Type != lexer.TOKEN_RBRACE && p.current.Type != lexer.TOKEN_EOF {
		if stmt := p.parseStatement(); stmt != nil {
			node.Statements = append(node.Statements, stmt)
		}
	}

	p.expect(lexer.TOKEN_RBRACE)
	return node
}

func (p *Parser) parseControl() *ast.ControlNode {
	node := &ast.ControlNode{Line: p.current.Line}
	p.expect(lexer.TOKEN_CONTROL)
	p.expect(lexer.TOKEN_LBRACE)

	for p.current.Type != lexer.TOKEN_RBRACE && p.current.Type != lexer.TOKEN_EOF {
		if stmt := p.parseStatement(); stmt != nil {
			node.Statements = append(node.Statements, stmt)
		}
	}

	p.expect(lexer.TOKEN_RBRACE)
	return node
}

func (p *Parser) parseStatement() ast.Node {
	switch p.current.Type {
	case lexer.TOKEN_IF:
		return p.parseIf()
	case lexer.TOKEN_OUTPUT:
		return p.parseOutputCall()
	case lexer.TOKEN_IDENT:
		return p.parseAssign()
	default:
		p.advance()
		return nil
	}
}

func (p *Parser) parseAssign() *ast.AssignStatement {
	node := &ast.AssignStatement{Line: p.current.Line}
	node.Name = p.current.Literal
	p.advance()
	p.expect(lexer.TOKEN_ASSIGN)
	node.Value = p.current.Literal
	p.advance()
	return node
}

func (p *Parser) parseOutputCall() *ast.OutputCall {
	node := &ast.OutputCall{Line: p.current.Line}
	p.expect(lexer.TOKEN_OUTPUT)
	p.expect(lexer.TOKEN_LPAREN)
	node.Relay = p.current.Literal
	p.advance()
	p.expect(lexer.TOKEN_COMMA)
	node.State = p.current.Literal
	p.advance()
	p.expect(lexer.TOKEN_RPAREN)
	return node
}

func (p *Parser) parseComparisonOperator() string {
	op := p.current.Literal
	switch p.current.Type {
	case lexer.TOKEN_GTE, lexer.TOKEN_GT, lexer.TOKEN_LTE, lexer.TOKEN_LT, lexer.TOKEN_NEQ, lexer.TOKEN_EQ:
		p.advance()
		return op
	default:
		p.Errors = append(p.Errors,
			fmt.Sprintf("line %d: expected comparison operator but got %s (%q)",
				p.current.Line, p.current.Type, p.current.Literal))
		p.advance()
		return op
	}
}

func (p *Parser) parseIf() ast.Node {
	line := p.current.Line
	p.expect(lexer.TOKEN_IF)

	leftVar := p.current.Literal
	p.advance()
	leftOp := p.parseComparisonOperator()
	leftVal := p.current.Literal
	p.advance()

	// AND compound condition
	if p.current.Type == lexer.TOKEN_AND {
		p.advance()
		rightVar := p.current.Literal
		p.advance()
		rightOp := p.parseComparisonOperator()
		rightVal := p.current.Literal
		p.advance()

		p.expect(lexer.TOKEN_COLON)
		then := p.parseAssign()

		p.expect(lexer.TOKEN_ELSE)
		p.expect(lexer.TOKEN_COLON)
		els := p.parseAssign()

		return &ast.IfElseStatement{
			Line:     line,
			LeftVar:  leftVar,
			LeftOp:   leftOp,
			LeftVal:  leftVal,
			RightVar: rightVar,
			RightOp:  rightOp,
			RightVal: rightVal,
			Then:     then,
			Else:     els,
		}
	}

	// OR compound condition
	if p.current.Type == lexer.TOKEN_OR {
		p.advance()
		rightVar := p.current.Literal
		p.advance()
		rightOp := p.parseComparisonOperator()
		rightVal := p.current.Literal
		p.advance()

		p.expect(lexer.TOKEN_COLON)
		then := p.parseAssign()

		if p.current.Type == lexer.TOKEN_ELSE {
			p.expect(lexer.TOKEN_ELSE)
			p.expect(lexer.TOKEN_COLON)
			els := p.parseAssign()
			return &ast.IfOrElseStatement{
				Line:     line,
				LeftVar:  leftVar,
				LeftOp:   leftOp,
				LeftVal:  leftVal,
				RightVar: rightVar,
				RightOp:  rightOp,
				RightVal: rightVal,
				Then:     then,
				Else:     els,
			}
		}

		return &ast.IfOrStatement{
			Line:     line,
			LeftVar:  leftVar,
			LeftOp:   leftOp,
			LeftVal:  leftVal,
			RightVar: rightVar,
			RightOp:  rightOp,
			RightVal: rightVal,
			Then:     then,
		}
	}

	// simple if
	p.expect(lexer.TOKEN_COLON)
	then := p.parseAssign()

	return &ast.IfStatement{
		Line:     line,
		Left:     leftVar,
		Operator: leftOp,
		Right:    leftVal,
		Then:     then,
	}
}
