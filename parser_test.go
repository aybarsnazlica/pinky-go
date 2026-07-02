package pinky

import "testing"

func TestParserCursorOperationsMatchOriginalBehavior(t *testing.T) {
	parser := newParserAt([]Token{
		{Kind: TokenIDENTIFIER, Lexeme: "alpha", Line: 1},
		{Kind: TokenINTEGER, Lexeme: "42", Line: 2},
		{Kind: TokenPRINT, Lexeme: "print", Line: 3},
	}, 0)

	if got := parser.advance(); got != (Token{Kind: TokenIDENTIFIER, Lexeme: "alpha", Line: 1}) {
		t.Fatalf("advance() = %+v", got)
	}
	if parser.current != 1 {
		t.Fatalf("current = %d", parser.current)
	}
	if got := parser.peek(); got != (Token{Kind: TokenINTEGER, Lexeme: "42", Line: 2}) {
		t.Fatalf("peek() = %+v", got)
	}
	if !parser.isNext(TokenINTEGER) {
		t.Fatal("isNext(TokenINTEGER) = false")
	}
	if parser.isNext(TokenIDENTIFIER) {
		t.Fatal("isNext(TokenIDENTIFIER) = true")
	}
	if got, err := parser.expect(TokenINTEGER); err != nil || got != (Token{Kind: TokenINTEGER, Lexeme: "42", Line: 2}) {
		t.Fatalf("expect(TokenINTEGER) = %+v, %v", got, err)
	}
	if got := parser.peek(); got != (Token{Kind: TokenPRINT, Lexeme: "print", Line: 3}) {
		t.Fatalf("peek() after expect = %+v", got)
	}
}

func TestParserErrorsPreserveMessages(t *testing.T) {
	mismatch := newParserAt([]Token{
		{Kind: TokenIDENTIFIER, Lexeme: "alpha", Line: 1},
		{Kind: TokenINTEGER, Lexeme: "42", Line: 2},
		{Kind: TokenPRINT, Lexeme: "print", Line: 3},
	}, 1)
	mismatchError := expectParseError(t, func() error {
		_, err := mismatch.expect(TokenIDENTIFIER)
		return err
	})
	if mismatchError.Line != 2 {
		t.Fatalf("line = %d", mismatchError.Line)
	}
	if mismatchError.MessageText != "Expected 'TOK_IDENTIFIER', found '42'." {
		t.Fatalf("message = %q", mismatchError.MessageText)
	}

	end := newParserAt([]Token{
		{Kind: TokenIDENTIFIER, Lexeme: "alpha", Line: 1},
		{Kind: TokenINTEGER, Lexeme: "42", Line: 2},
		{Kind: TokenPRINT, Lexeme: "print", Line: 3},
	}, 3)
	endError := expectParseError(t, func() error {
		_, err := end.expect(TokenPRINT)
		return err
	})
	if endError.MessageText != "Found 'print' at the end of parsing" {
		t.Fatalf("message = %q", endError.MessageText)
	}
}

func TestParserBuildsExpectedExpressionsAndStatements(t *testing.T) {
	primary, err := parsePrimary([]Token{{Kind: TokenINTEGER, Lexeme: "42", Line: 1}})
	if err != nil {
		t.Fatalf("parsePrimary() error = %v", err)
	}
	if got := primary.String(); got != "Integer[42]" {
		t.Fatalf("primary = %q", got)
	}

	grouping, err := parsePrimary([]Token{
		{Kind: TokenLPAREN, Lexeme: "(", Line: 1},
		{Kind: TokenINTEGER, Lexeme: "1", Line: 1},
		{Kind: TokenPLUS, Lexeme: "+", Line: 1},
		{Kind: TokenINTEGER, Lexeme: "2", Line: 1},
		{Kind: TokenRPAREN, Lexeme: ")", Line: 1},
	})
	if err != nil {
		t.Fatalf("grouping parse error = %v", err)
	}
	if got := grouping.String(); got != `Grouping(BinOp("+", Integer[1], Integer[2]))` {
		t.Fatalf("grouping = %q", got)
	}

	addition, err := parseExpr([]Token{
		{Kind: TokenINTEGER, Lexeme: "1", Line: 1},
		{Kind: TokenPLUS, Lexeme: "+", Line: 1},
		{Kind: TokenINTEGER, Lexeme: "2", Line: 1},
		{Kind: TokenSTAR, Lexeme: "*", Line: 1},
		{Kind: TokenINTEGER, Lexeme: "3", Line: 1},
	})
	if err != nil {
		t.Fatalf("parseExpr() error = %v", err)
	}
	if got := addition.String(); got != `BinOp("+", Integer[1], BinOp("*", Integer[2], Integer[3]))` {
		t.Fatalf("addition = %q", got)
	}

	printStmt, err := parseStmt([]Token{{Kind: TokenPRINT, Lexeme: "print", Line: 1}, {Kind: TokenSTRING, Lexeme: `"x"`, Line: 1}})
	if err != nil {
		t.Fatalf("parseStmt(print) error = %v", err)
	}
	if got := printStmt.String(); got != `PrintStmt(String[x], end="")` {
		t.Fatalf("print stmt = %q", got)
	}

	functionDecl, err := parseStmt([]Token{
		{Kind: TokenFUNC, Lexeme: "func", Line: 12},
		{Kind: TokenIDENTIFIER, Lexeme: "add", Line: 12},
		{Kind: TokenLPAREN, Lexeme: "(", Line: 12},
		{Kind: TokenIDENTIFIER, Lexeme: "a", Line: 12},
		{Kind: TokenCOMMA, Lexeme: ",", Line: 12},
		{Kind: TokenIDENTIFIER, Lexeme: "b", Line: 12},
		{Kind: TokenRPAREN, Lexeme: ")", Line: 12},
		{Kind: TokenRET, Lexeme: "ret", Line: 13},
		{Kind: TokenIDENTIFIER, Lexeme: "a", Line: 13},
		{Kind: TokenPLUS, Lexeme: "+", Line: 13},
		{Kind: TokenIDENTIFIER, Lexeme: "b", Line: 13},
		{Kind: TokenEND, Lexeme: "end", Line: 14},
	})
	if err != nil {
		t.Fatalf("parseStmt(func) error = %v", err)
	}
	if got := functionDecl.String(); got != `FuncDecl("add", [Param["a"], Param["b"]], Stmts([RetStmt[BinOp("+", Identifier["a"], Identifier["b"])]]))` {
		t.Fatalf("func decl = %q", got)
	}
}

func TestParserRejectsBrokenGroupingAndBareExpressionStatements(t *testing.T) {
	missingParen := newParserAt([]Token{{Kind: TokenLPAREN, Lexeme: "(", Line: 1}, {Kind: TokenINTEGER, Lexeme: "1", Line: 1}}, 0)
	missingParenError := expectParseError(t, func() error {
		_, err := missingParen.primary()
		return err
	})
	if missingParenError.MessageText != `Error: ")" expected.` {
		t.Fatalf("message = %q", missingParenError.MessageText)
	}

	bareExpression := newParserAt([]Token{{Kind: TokenIDENTIFIER, Lexeme: "x", Line: 1}}, 0)
	bareExpressionError := expectParseError(t, func() error {
		_, err := bareExpression.stmt()
		return err
	})
	if bareExpressionError.MessageText != "Expected assignment or function call" {
		t.Fatalf("message = %q", bareExpressionError.MessageText)
	}
}

func parsePrimary(tokens []Token) (Expr, error) {
	return newParserAt(tokens, 0).primary()
}

func parseExpr(tokens []Token) (Expr, error) {
	return newParserAt(tokens, 0).expr()
}

func parseStmt(tokens []Token) (Stmt, error) {
	return newParserAt(tokens, 0).stmt()
}

func expectParseError(t *testing.T, fn func() error) *ParseError {
	t.Helper()
	err := fn()
	if err == nil {
		t.Fatal("expected ParseError")
	}
	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	return parseErr
}
