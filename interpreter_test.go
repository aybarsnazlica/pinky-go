package pinky

import (
	"math"
	"testing"
)

func TestInterpretEvaluatesLiteralsUnaryBinaryAndGrouping(t *testing.T) {
	interpreter := NewInterpreter(nil)
	env := NewEnvironment()

	assertNumber(t, evalMust(t, interpreter, &IntegerLiteral{Value: 7, line: 1}, env), 7)
	assertNumber(t, evalMust(t, interpreter, &FloatLiteral{Value: 2.5, line: 1}, env), 2.5)
	assertString(t, evalMust(t, interpreter, &StringLiteral{Value: "hi", line: 1}, env), "hi")
	assertBool(t, evalMust(t, interpreter, &BoolLiteral{Value: true, line: 1}, env), true)
	assertNumber(t, evalMust(t, interpreter, &GroupingExpr{Value: &IntegerLiteral{Value: 4, line: 1}, line: 1}, env), 4)

	negate := &UnaryExpr{Op: Token{Kind: TokenMINUS, Lexeme: "-", Line: 1}, Operand: &IntegerLiteral{Value: 3, line: 1}, line: 1}
	assertNumber(t, evalMust(t, interpreter, negate, env), -3)

	add := &BinaryExpr{Op: Token{Kind: TokenPLUS, Lexeme: "+", Line: 1}, Left: &IntegerLiteral{Value: 2, line: 1}, Right: &FloatLiteral{Value: 3.5, line: 1}, line: 1}
	assertNumber(t, evalMust(t, interpreter, add, env), 5.5)

	concat := &BinaryExpr{Op: Token{Kind: TokenPLUS, Lexeme: "+", Line: 1}, Left: &StringLiteral{Value: "value=", line: 1}, Right: &IntegerLiteral{Value: 5, line: 1}, line: 1}
	assertString(t, evalMust(t, interpreter, concat, env), "value=5")

	modulo := &BinaryExpr{Op: Token{Kind: TokenMOD, Lexeme: "%", Line: 1}, Left: &IntegerLiteral{Value: -90, line: 1}, Right: &IntegerLiteral{Value: 360, line: 1}, line: 1}
	assertNumber(t, evalMust(t, interpreter, modulo, env), 270)

	equalNegativeZero := &BinaryExpr{Op: Token{Kind: TokenEQEQ, Lexeme: "==", Line: 1}, Left: &FloatLiteral{Value: mathCopysignZero(-1), line: 1}, Right: &IntegerLiteral{Value: 0, line: 1}, line: 1}
	assertBool(t, evalMust(t, interpreter, equalNegativeZero, env), true)

	notEqualNegativeZero := &BinaryExpr{Op: Token{Kind: TokenNE, Lexeme: "~=", Line: 1}, Left: &FloatLiteral{Value: mathCopysignZero(-1), line: 1}, Right: &IntegerLiteral{Value: 0, line: 1}, line: 1}
	assertBool(t, evalMust(t, interpreter, notEqualNegativeZero, env), false)
}

func TestInterpretHandlesVariablesControlFlowAndFunctions(t *testing.T) {
	interpreter := NewInterpreter(nil)
	env := NewEnvironment()
	env.SetLocal("sum", NumberValue(0))

	if err := interpreter.Execute(&AssignmentStmt{Left: &IdentifierExpr{Name: "sum", line: 1}, Right: &IntegerLiteral{Value: 4, line: 1}, line: 1}, env); err != nil {
		t.Fatalf("assignment error = %v", err)
	}
	assertNumber(t, evalMust(t, interpreter, &IdentifierExpr{Name: "sum", line: 1}, env), 4)

	env.SetLocal("count", NumberValue(3))
	whileStmt := &WhileStmt{
		Test: &BinaryExpr{Op: Token{Kind: TokenGT, Lexeme: ">", Line: 2}, Left: &IdentifierExpr{Name: "count", line: 2}, Right: &IntegerLiteral{Value: 0, line: 2}, line: 2},
		BodyStmts: &Program{Statements: []Stmt{
			&AssignmentStmt{Left: &IdentifierExpr{Name: "count", line: 2}, Right: &BinaryExpr{Op: Token{Kind: TokenMINUS, Lexeme: "-", Line: 2}, Left: &IdentifierExpr{Name: "count", line: 2}, Right: &IntegerLiteral{Value: 1, line: 2}, line: 2}, line: 2},
		}, line: 2},
		line: 2,
	}
	if err := interpreter.Execute(whileStmt, env); err != nil {
		t.Fatalf("while error = %v", err)
	}
	assertNumber(t, evalMust(t, interpreter, &IdentifierExpr{Name: "count", line: 2}, env), 0)

	forStmt := &ForStmt{
		Ident: &IdentifierExpr{Name: "i", line: 3},
		Start: &IntegerLiteral{Value: 1, line: 3},
		End:   &IntegerLiteral{Value: 3, line: 3},
		BodyStmts: &Program{Statements: []Stmt{
			&AssignmentStmt{Left: &IdentifierExpr{Name: "sum", line: 3}, Right: &BinaryExpr{Op: Token{Kind: TokenPLUS, Lexeme: "+", Line: 3}, Left: &IdentifierExpr{Name: "sum", line: 3}, Right: &IdentifierExpr{Name: "i", line: 3}, line: 3}, line: 3},
		}, line: 3},
		line: 3,
	}
	if err := interpreter.Execute(forStmt, env); err != nil {
		t.Fatalf("for error = %v", err)
	}
	assertNumber(t, evalMust(t, interpreter, &IdentifierExpr{Name: "sum", line: 3}, env), 10)

	functionDecl := &FunctionDecl{
		Name:   "add_base",
		Params: []*Param{{Name: "x", line: 1}},
		BodyStmts: &Program{Statements: []Stmt{
			&ReturnStmt{Value: &BinaryExpr{Op: Token{Kind: TokenPLUS, Lexeme: "+", Line: 2}, Left: &IdentifierExpr{Name: "x", line: 2}, Right: &IdentifierExpr{Name: "base", line: 2}, line: 2}, line: 2},
		}, line: 2},
		line: 1,
	}
	env.SetLocal("base", NumberValue(10))
	if err := interpreter.Execute(functionDecl, env); err != nil {
		t.Fatalf("func decl error = %v", err)
	}
	assertNumber(t, evalMust(t, interpreter, &FunctionCallExpr{Name: "add_base", Args: []Expr{&IntegerLiteral{Value: 5, line: 3}}, line: 3}, env), 15)
}

func TestInterpretPrintStatementsAndFunctionCallStatements(t *testing.T) {
	output := ""
	interpreter := NewInterpreter(func(text string) {
		output += text
	})
	env := NewEnvironment()

	if err := interpreter.Execute(&PrintStmt{Value: &StringLiteral{Value: `hello\nworld`, line: 1}, End: "", line: 1}, env); err != nil {
		t.Fatalf("print error = %v", err)
	}
	if output != "hello\nworld" {
		t.Fatalf("output = %q", output)
	}

	functionDecl := &FunctionDecl{Name: "f", Params: []*Param{}, BodyStmts: &Program{Statements: []Stmt{&ReturnStmt{Value: &IntegerLiteral{Value: 8, line: 1}, line: 1}}, line: 1}, line: 1}
	if err := interpreter.Execute(functionDecl, env); err != nil {
		t.Fatalf("func decl error = %v", err)
	}
	result, err := interpreter.Interpret(&FunctionCallStmt{ExprValue: &FunctionCallExpr{Name: "f", Args: []Expr{}, line: 2}}, env)
	if err != nil {
		t.Fatalf("interpret stmt error = %v", err)
	}
	if result != nil {
		t.Fatalf("result = %+v", result)
	}

	err = interpreter.Execute(&ReturnStmt{Value: &IntegerLiteral{Value: 1, line: 3}, line: 3}, env)
	if err == nil {
		t.Fatal("expected return signal")
	}
	signal, ok := err.(*returnSignal)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	assertNumber(t, signal.value, 1)
}

func TestExecuteFunctionCallStatementDoesNotRequireReturnValue(t *testing.T) {
	interpreter := NewInterpreter(nil)
	env := NewEnvironment()

	if err := interpreter.Execute(&FunctionDecl{Name: "f", Params: []*Param{}, BodyStmts: &Program{Statements: []Stmt{}, line: 1}, line: 1}, env); err != nil {
		t.Fatalf("func decl error = %v", err)
	}
	if err := interpreter.Execute(&FunctionCallStmt{ExprValue: &FunctionCallExpr{Name: "f", Args: []Expr{}, line: 2}}, env); err != nil {
		t.Fatalf("func call stmt error = %v", err)
	}
}

func TestInterpretFunctionCallExpressionCanStillReturnEmptyFromWrapper(t *testing.T) {
	interpreter := NewInterpreter(nil)
	env := NewEnvironment()

	if err := interpreter.Execute(&FunctionDecl{Name: "f", Params: []*Param{}, BodyStmts: &Program{Statements: []Stmt{}, line: 1}, line: 1}, env); err != nil {
		t.Fatalf("func decl error = %v", err)
	}

	result, err := interpreter.Interpret(&FunctionCallExpr{Name: "f", Args: []Expr{}, line: 2}, env)
	if err != nil {
		t.Fatalf("interpret expr error = %v", err)
	}
	if result != nil {
		t.Fatalf("result = %+v", result)
	}

	_, err = interpreter.Evaluate(&FunctionCallExpr{Name: "f", Args: []Expr{}, line: 2}, env)
	runtimeErr := expectRuntimeError(t, err)
	if runtimeErr.MessageText != "Function 'f' did not return a value." {
		t.Fatalf("message = %q", runtimeErr.MessageText)
	}
}

func TestInterpretThrowsRuntimeErrorsAndBuildsGlobalEnv(t *testing.T) {
	interpreter := NewInterpreter(nil)
	env := NewEnvironment()

	_, err := interpreter.Evaluate(&IdentifierExpr{Name: "missing", line: 1}, env)
	runtimeErr := expectRuntimeError(t, err)
	if runtimeErr.MessageText != "Undeclared identifier 'missing'" {
		t.Fatalf("message = %q", runtimeErr.MessageText)
	}

	_, err = interpreter.Evaluate(&BinaryExpr{Op: Token{Kind: TokenSLASH, Lexeme: "/", Line: 2}, Left: &IntegerLiteral{Value: 4, line: 2}, Right: &IntegerLiteral{Value: 0, line: 2}, line: 2}, env)
	runtimeErr = expectRuntimeError(t, err)
	if runtimeErr.MessageText != "Division by zero." {
		t.Fatalf("message = %q", runtimeErr.MessageText)
	}

	_, err = interpreter.Interpret(&FunctionCallExpr{Name: "missing_func", Args: []Expr{}, line: 3}, env)
	runtimeErr = expectRuntimeError(t, err)
	if runtimeErr.MessageText != "Function 'missing_func' not declared." {
		t.Fatalf("message = %q", runtimeErr.MessageText)
	}

	global, err := interpreter.InterpretAST(&Program{Statements: []Stmt{&LocalAssignmentStmt{Left: &IdentifierExpr{Name: "x", line: 1}, Right: &IntegerLiteral{Value: 6, line: 1}, line: 1}}, line: 1})
	if err != nil {
		t.Fatalf("InterpretAST() error = %v", err)
	}
	if value, ok := global.GetVar("x"); !ok || value != NumberValue(6) {
		t.Fatalf("global.GetVar(x) = %+v, %v", value, ok)
	}
}

func mustValue(t *testing.T, value RuntimeValue, err error) RuntimeValue {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return value
}

func evalMust(t *testing.T, interpreter *Interpreter, expr Expr, env *Environment) RuntimeValue {
	t.Helper()
	value, err := interpreter.Evaluate(expr, env)
	return mustValue(t, value, err)
}

func assertNumber(t *testing.T, value RuntimeValue, expected float64) {
	t.Helper()
	if value.Type != RuntimeTypeNumber || value.Number != expected {
		t.Fatalf("number = %+v, want %v", value, expected)
	}
}

func assertString(t *testing.T, value RuntimeValue, expected string) {
	t.Helper()
	if value.Type != RuntimeTypeString || value.String != expected {
		t.Fatalf("string = %+v, want %q", value, expected)
	}
}

func assertBool(t *testing.T, value RuntimeValue, expected bool) {
	t.Helper()
	if value.Type != RuntimeTypeBool || value.Bool != expected {
		t.Fatalf("bool = %+v, want %v", value, expected)
	}
}

func expectRuntimeError(t *testing.T, err error) *RuntimeError {
	t.Helper()
	if err == nil {
		t.Fatal("expected RuntimeError")
	}
	runtimeErr, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	return runtimeErr
}

func mathCopysignZero(sign float64) float64 {
	return math.Copysign(0, sign)
}
