package pinky

import (
	"strconv"
	"strings"
)

type Parser struct {
	tokens  []Token
	current int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: append([]Token(nil), tokens...)}
}

func newParserAt(tokens []Token, current int) *Parser {
	return &Parser{tokens: append([]Token(nil), tokens...), current: current}
}

func (p *Parser) advance() Token {
	token := p.tokens[p.current]
	p.current++
	return token
}

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

func (p *Parser) isNext(expectedType TokenType) bool {
	return p.current < len(p.tokens) && p.tokens[p.current].Kind == expectedType
}

func (p *Parser) expect(expectedType TokenType) (Token, error) {
	if p.current >= len(p.tokens) {
		previous := p.previousToken()
		return Token{}, &ParseError{Line: previous.Line, MessageText: "Found " + quote(previous.Lexeme) + " at the end of parsing"}
	}

	if p.peek().Kind == expectedType {
		return p.advance(), nil
	}

	token := p.peek()
	return Token{}, &ParseError{Line: token.Line, MessageText: "Expected " + quote(tokenDebugName(expectedType)) + ", found " + quote(token.Lexeme) + "."}
}

func (p *Parser) previousToken() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) match(expectedType TokenType) bool {
	if p.current >= len(p.tokens) {
		return false
	}
	if p.peek().Kind != expectedType {
		return false
	}
	p.current++
	return true
}

func (p *Parser) primary() (Expr, error) {
	if p.match(TokenINTEGER) {
		previous := p.previousToken()
		value, _ := strconv.Atoi(previous.Lexeme)
		return &IntegerLiteral{Value: value, line: previous.Line}, nil
	}
	if p.match(TokenFLOAT) {
		previous := p.previousToken()
		value, _ := strconv.ParseFloat(previous.Lexeme, 64)
		return &FloatLiteral{Value: value, line: previous.Line}, nil
	}
	if p.match(TokenTRUE) {
		return &BoolLiteral{Value: true, line: p.previousToken().Line}, nil
	}
	if p.match(TokenFALSE) {
		return &BoolLiteral{Value: false, line: p.previousToken().Line}, nil
	}
	if p.match(TokenSTRING) {
		lexeme := p.previousToken().Lexeme
		return &StringLiteral{Value: lexeme[1 : len(lexeme)-1], line: p.previousToken().Line}, nil
	}
	if p.match(TokenLPAREN) {
		expression, err := p.expr()
		if err != nil {
			return nil, err
		}
		if !p.match(TokenRPAREN) {
			return nil, &ParseError{Line: p.previousToken().Line, MessageText: `Error: ")" expected.`}
		}
		return &GroupingExpr{Value: expression, line: p.previousToken().Line}, nil
	}

	identifier, err := p.expect(TokenIDENTIFIER)
	if err != nil {
		return nil, err
	}
	if p.match(TokenLPAREN) {
		args, err := p.args()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRPAREN); err != nil {
			return nil, err
		}
		return &FunctionCallExpr{Name: identifier.Lexeme, Args: args, line: p.previousToken().Line}, nil
	}
	return &IdentifierExpr{Name: identifier.Lexeme, line: identifier.Line}, nil
}

func (p *Parser) exponent() (Expr, error) {
	expression, err := p.primary()
	if err != nil {
		return nil, err
	}
	for p.match(TokenCARET) {
		op := p.previousToken()
		right, err := p.exponent()
		if err != nil {
			return nil, err
		}
		expression = &BinaryExpr{Op: op, Left: expression, Right: right, line: op.Line}
	}
	return expression, nil
}

func (p *Parser) unary() (Expr, error) {
	if p.match(TokenNOT) || p.match(TokenMINUS) || p.match(TokenPLUS) {
		op := p.previousToken()
		operand, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: op, Operand: operand, line: op.Line}, nil
	}
	return p.exponent()
}

func (p *Parser) modulo() (Expr, error) {
	expression, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.match(TokenMOD) {
		op := p.previousToken()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expression = &BinaryExpr{Op: op, Left: expression, Right: right, line: op.Line}
	}
	return expression, nil
}

func (p *Parser) multiplication() (Expr, error) {
	expression, err := p.modulo()
	if err != nil {
		return nil, err
	}
	for p.match(TokenSTAR) || p.match(TokenSLASH) {
		op := p.previousToken()
		right, err := p.modulo()
		if err != nil {
			return nil, err
		}
		expression = &BinaryExpr{Op: op, Left: expression, Right: right, line: op.Line}
	}
	return expression, nil
}

func (p *Parser) addition() (Expr, error) {
	expression, err := p.multiplication()
	if err != nil {
		return nil, err
	}
	for p.match(TokenPLUS) || p.match(TokenMINUS) {
		op := p.previousToken()
		right, err := p.multiplication()
		if err != nil {
			return nil, err
		}
		expression = &BinaryExpr{Op: op, Left: expression, Right: right, line: op.Line}
	}
	return expression, nil
}

func (p *Parser) comparison() (Expr, error) {
	expression, err := p.addition()
	if err != nil {
		return nil, err
	}
	for p.match(TokenGT) || p.match(TokenGE) || p.match(TokenLT) || p.match(TokenLE) {
		op := p.previousToken()
		right, err := p.addition()
		if err != nil {
			return nil, err
		}
		expression = &BinaryExpr{Op: op, Left: expression, Right: right, line: op.Line}
	}
	return expression, nil
}

func (p *Parser) equality() (Expr, error) {
	expression, err := p.comparison()
	if err != nil {
		return nil, err
	}
	for p.match(TokenNE) || p.match(TokenEQEQ) {
		op := p.previousToken()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expression = &BinaryExpr{Op: op, Left: expression, Right: right, line: op.Line}
	}
	return expression, nil
}

func (p *Parser) logicalAnd() (Expr, error) {
	expression, err := p.equality()
	if err != nil {
		return nil, err
	}
	for p.match(TokenAND) {
		op := p.previousToken()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}
		expression = &LogicalExpr{Op: op, Left: expression, Right: right, line: op.Line}
	}
	return expression, nil
}

func (p *Parser) logicalOr() (Expr, error) {
	expression, err := p.logicalAnd()
	if err != nil {
		return nil, err
	}
	for p.match(TokenOR) {
		op := p.previousToken()
		right, err := p.logicalAnd()
		if err != nil {
			return nil, err
		}
		expression = &LogicalExpr{Op: op, Left: expression, Right: right, line: op.Line}
	}
	return expression, nil
}

func (p *Parser) expr() (Expr, error) {
	return p.logicalOr()
}

func (p *Parser) printStmt(end string) (Stmt, error) {
	if p.match(TokenPRINT) || p.match(TokenPRINTLN) {
		expression, err := p.expr()
		if err != nil {
			return nil, err
		}
		return &PrintStmt{Value: expression, End: end, line: p.previousToken().Line}, nil
	}
	return nil, nil
}

func (p *Parser) ifStmt() (Stmt, error) {
	if _, err := p.expect(TokenIF); err != nil {
		return nil, err
	}
	test, err := p.expr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenTHEN); err != nil {
		return nil, err
	}
	thenStmts, err := p.stmts()
	if err != nil {
		return nil, err
	}
	var elseStmts *Program
	if p.isNext(TokenELSE) {
		p.advance()
		elseStmts, err = p.stmts()
		if err != nil {
			return nil, err
		}
	}
	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	return &IfStmt{Test: test, ThenStmts: thenStmts, ElseStmts: elseStmts, line: p.previousToken().Line}, nil
}

func (p *Parser) whileStmt() (Stmt, error) {
	if _, err := p.expect(TokenWHILE); err != nil {
		return nil, err
	}
	test, err := p.expr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenDO); err != nil {
		return nil, err
	}
	bodyStmts, err := p.stmts()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	return &WhileStmt{Test: test, BodyStmts: bodyStmts, line: p.previousToken().Line}, nil
}

func (p *Parser) forStmt() (Stmt, error) {
	if _, err := p.expect(TokenFOR); err != nil {
		return nil, err
	}
	identifier, err := p.expect(TokenIDENTIFIER)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenASSIGN); err != nil {
		return nil, err
	}
	start, err := p.expr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenCOMMA); err != nil {
		return nil, err
	}
	end, err := p.expr()
	if err != nil {
		return nil, err
	}
	var step Expr
	if p.isNext(TokenCOMMA) {
		p.advance()
		step, err = p.expr()
		if err != nil {
			return nil, err
		}
	}
	if _, err := p.expect(TokenDO); err != nil {
		return nil, err
	}
	bodyStmts, err := p.stmts()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	return &ForStmt{Ident: &IdentifierExpr{Name: identifier.Lexeme, line: identifier.Line}, Start: start, End: end, Step: step, BodyStmts: bodyStmts, line: p.previousToken().Line}, nil
}

func (p *Parser) args() ([]Expr, error) {
	args := make([]Expr, 0)
	for !p.isNext(TokenRPAREN) {
		expression, err := p.expr()
		if err != nil {
			return nil, err
		}
		args = append(args, expression)
		if !p.isNext(TokenRPAREN) {
			if _, err := p.expect(TokenCOMMA); err != nil {
				return nil, err
			}
		}
	}
	return args, nil
}

func (p *Parser) params() ([]*Param, error) {
	params := make([]*Param, 0)
	numParams := 0
	for !p.isNext(TokenRPAREN) {
		name, err := p.expect(TokenIDENTIFIER)
		if err != nil {
			return nil, err
		}
		numParams++
		if numParams > 255 {
			return nil, &ParseError{Line: name.Line, MessageText: "Functions cannot have more than 255 parameters."}
		}
		params = append(params, &Param{Name: name.Lexeme, line: p.previousToken().Line})
		if !p.isNext(TokenRPAREN) {
			if _, err := p.expect(TokenCOMMA); err != nil {
				return nil, err
			}
		}
	}
	return params, nil
}

func (p *Parser) funcDecl() (Stmt, error) {
	if _, err := p.expect(TokenFUNC); err != nil {
		return nil, err
	}
	name, err := p.expect(TokenIDENTIFIER)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenLPAREN); err != nil {
		return nil, err
	}
	params, err := p.params()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenRPAREN); err != nil {
		return nil, err
	}
	bodyStmts, err := p.stmts()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenEND); err != nil {
		return nil, err
	}
	return &FunctionDecl{Name: name.Lexeme, Params: params, BodyStmts: bodyStmts, line: name.Line}, nil
}

func (p *Parser) retStmt() (Stmt, error) {
	if _, err := p.expect(TokenRET); err != nil {
		return nil, err
	}
	expression, err := p.expr()
	if err != nil {
		return nil, err
	}
	return &ReturnStmt{Value: expression, line: p.previousToken().Line}, nil
}

func (p *Parser) localAssign() (Stmt, error) {
	if _, err := p.expect(TokenLOCAL); err != nil {
		return nil, err
	}
	left, err := p.expr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenASSIGN); err != nil {
		return nil, err
	}
	right, err := p.expr()
	if err != nil {
		return nil, err
	}
	return &LocalAssignmentStmt{Left: left, Right: right, line: p.previousToken().Line}, nil
}

func (p *Parser) stmt() (Stmt, error) {
	if p.peek().Kind == TokenPRINT {
		return p.printStmt("")
	}
	if p.peek().Kind == TokenPRINTLN {
		return p.printStmt("\n")
	}
	if p.peek().Kind == TokenIF {
		return p.ifStmt()
	}
	if p.peek().Kind == TokenWHILE {
		return p.whileStmt()
	}
	if p.peek().Kind == TokenFOR {
		return p.forStmt()
	}
	if p.peek().Kind == TokenFUNC {
		return p.funcDecl()
	}
	if p.peek().Kind == TokenRET {
		return p.retStmt()
	}
	if p.peek().Kind == TokenLOCAL {
		return p.localAssign()
	}

	left, err := p.expr()
	if err != nil {
		return nil, err
	}
	if p.match(TokenASSIGN) {
		right, err := p.expr()
		if err != nil {
			return nil, err
		}
		return &AssignmentStmt{Left: left, Right: right, line: p.previousToken().Line}, nil
	}
	functionCallExpr, err := expectFunctionCallExpr(left, p.previousToken().Line)
	if err != nil {
		return nil, err
	}
	return &FunctionCallStmt{ExprValue: functionCallExpr}, nil
}

func (p *Parser) stmts() (*Program, error) {
	statements := make([]Stmt, 0)
	for p.current < len(p.tokens) && !p.isNext(TokenELSE) && !p.isNext(TokenEND) {
		statement, err := p.stmt()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}
	return &Program{Statements: statements, line: lastMeaningfulLine(p.tokens, p.current)}, nil
}

func (p *Parser) program() (*Program, error) {
	return p.stmts()
}

func (p *Parser) Parse() (*Program, error) {
	return p.program()
}

func lastMeaningfulLine(tokens []Token, current int) int {
	if current > 0 {
		return tokens[current-1].Line
	}
	if current < len(tokens) {
		return tokens[current].Line
	}
	return 0
}

func expectFunctionCallExpr(expr Expr, line int) (*FunctionCallExpr, error) {
	functionCallExpr, ok := expr.(*FunctionCallExpr)
	if ok {
		return functionCallExpr, nil
	}
	return nil, &ParseError{Line: line, MessageText: "Expected assignment or function call"}
}

func parserQuote(text string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return "'" + replacer.Replace(text) + "'"
}
