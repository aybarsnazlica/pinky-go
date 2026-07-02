package pinky

import "testing"

func TestLiteralNodesRenderExpectedStrings(t *testing.T) {
	if got := (&IntegerLiteral{Value: 17, line: 3}).String(); got != "Integer[17]" {
		t.Fatalf("integer = %q", got)
	}
	if got := (&FloatLiteral{Value: 3.5, line: 3}).String(); got != "Float[3.5]" {
		t.Fatalf("float = %q", got)
	}
	if got := (&BoolLiteral{Value: true, line: 3}).String(); got != "Bool[true]" {
		t.Fatalf("bool = %q", got)
	}
	if got := (&StringLiteral{Value: "hello", line: 3}).String(); got != "String[hello]" {
		t.Fatalf("string = %q", got)
	}
	if got := (&IdentifierExpr{Name: "value", line: 3}).String(); got != `Identifier["value"]` {
		t.Fatalf("identifier = %q", got)
	}
}

func TestExpressionAndStatementNodesPreserveASTStrings(t *testing.T) {
	negate := &UnaryExpr{Op: Token{Kind: TokenMINUS, Lexeme: "-", Line: 2}, Operand: &IntegerLiteral{Value: 17, line: 2}, line: 2}
	if got := negate.String(); got != `UnOp("-", Integer[17])` {
		t.Fatalf("unary = %q", got)
	}

	sum := &BinaryExpr{Op: Token{Kind: TokenPLUS, Lexeme: "+", Line: 2}, Left: &IdentifierExpr{Name: "x", line: 2}, Right: &IntegerLiteral{Value: 1, line: 2}, line: 2}
	if got := sum.String(); got != `BinOp("+", Identifier["x"], Integer[1])` {
		t.Fatalf("binary = %q", got)
	}

	logicalAnd := &LogicalExpr{Op: Token{Kind: TokenAND, Lexeme: "and", Line: 2}, Left: &BoolLiteral{Value: true, line: 2}, Right: &BoolLiteral{Value: false, line: 2}, line: 2}
	if got := logicalAnd.String(); got != `LogicalOp("and", Bool[true], Bool[false])` {
		t.Fatalf("logical = %q", got)
	}

	print := &PrintStmt{Value: &StringLiteral{Value: "hello", line: 5}, End: "\n", line: 5}
	if got := print.String(); got != `PrintStmt(String[hello], end="\n")` {
		t.Fatalf("print = %q", got)
	}

	ifStmt := &IfStmt{
		Test:      &BoolLiteral{Value: true, line: 5},
		ThenStmts: &Program{Statements: []Stmt{&PrintStmt{Value: &StringLiteral{Value: "then", line: 6}, End: "\n", line: 6}}, line: 6},
		ElseStmts: &Program{Statements: []Stmt{&PrintStmt{Value: &StringLiteral{Value: "else", line: 7}, End: "", line: 7}}, line: 7},
		line:      5,
	}
	if got := ifStmt.String(); got != `IfStmt(Bool[true], then:Stmts([PrintStmt(String[then], end="\n")]), else:Stmts([PrintStmt(String[else], end="")]))` {
		t.Fatalf("if stmt = %q", got)
	}
}

func TestFunctionNodesRenderExpectedStrings(t *testing.T) {
	functionDecl := &FunctionDecl{Name: "greet", Params: []*Param{{Name: "name", line: 13}}, BodyStmts: &Program{Statements: []Stmt{&ReturnStmt{Value: &IdentifierExpr{Name: "name", line: 13}, line: 13}}, line: 13}, line: 13}
	if got := functionDecl.String(); got != `FuncDecl("greet", [Param["name"]], Stmts([RetStmt[Identifier["name"]]]))` {
		t.Fatalf("func decl = %q", got)
	}

	functionCallExpr := &FunctionCallExpr{Name: "sum", Args: []Expr{&IntegerLiteral{Value: 1, line: 15}, &IdentifierExpr{Name: "value", line: 15}}, line: 15}
	if got := functionCallExpr.String(); got != `FuncCall("sum", [Integer[1], Identifier["value"]])` {
		t.Fatalf("func call expr = %q", got)
	}

	functionCallStmt := &FunctionCallStmt{ExprValue: functionCallExpr}
	if got := functionCallStmt.String(); got != `FuncCallStmt(FuncCall("sum", [Integer[1], Identifier["value"]]))` {
		t.Fatalf("func call stmt = %q", got)
	}
}
