package pinky

import "testing"

func TestTokenDebugString(t *testing.T) {
	token := Token{Kind: TokenIDENTIFIER, Lexeme: "hello", Line: 7}
	if got := token.DebugString(); got != `(TOK_IDENTIFIER, "hello", 7)` {
		t.Fatalf("DebugString() = %q", got)
	}
}

func TestLexerCursorHelpers(t *testing.T) {
	lexer := NewLexer("abcd")
	if got := lexer.advance(); got != 'a' {
		t.Fatalf("advance() = %q", got)
	}
	if got := lexer.peek(); got != 'b' {
		t.Fatalf("peek() = %q", got)
	}
	if got := lexer.lookahead(1); got != 'c' {
		t.Fatalf("lookahead(1) = %q", got)
	}
	if !lexer.match('b') {
		t.Fatal("match('b') = false")
	}
	if lexer.match('z') {
		t.Fatal("match('z') = true")
	}
}

func TestHandleNumberCreatesIntegerAndFloatTokens(t *testing.T) {
	integerLexer := makePositionedLexer("123", 1, 1)
	integerLexer.handleNumber()
	assertTokensEqual(t, integerLexer.tokens, []Token{{Kind: TokenINTEGER, Lexeme: "123", Line: 1}})

	floatLexer := makePositionedLexer("12.5", 1, 1)
	floatLexer.handleNumber()
	assertTokensEqual(t, floatLexer.tokens, []Token{{Kind: TokenFLOAT, Lexeme: "12.5", Line: 1}})
}

func TestHandleStringAndIdentifierRespectKeywords(t *testing.T) {
	stringLexer := makePositionedLexer(`"abc"`, 1, 1)
	if err := stringLexer.handleString('"'); err != nil {
		t.Fatalf("handleString() error = %v", err)
	}
	assertTokensEqual(t, stringLexer.tokens, []Token{{Kind: TokenSTRING, Lexeme: `"abc"`, Line: 1}})

	keywordLexer := makePositionedLexer("while", 1, 1)
	keywordLexer.handleIdentifier()
	if keywordLexer.tokens[0].Kind != TokenWHILE {
		t.Fatalf("keyword kind = %v", keywordLexer.tokens[0].Kind)
	}

	identifierLexer := makePositionedLexer("abc_12", 1, 1)
	identifierLexer.handleIdentifier()
	assertTokensEqual(t, identifierLexer.tokens, []Token{{Kind: TokenIDENTIFIER, Lexeme: "abc_12", Line: 1}})
}

func TestTokenizeHandlesOperatorsKeywordsStringsAndComments(t *testing.T) {
	operators := NewLexer("()+-*/^%?:;= == ~= <= >= < > :=")
	tokens, err := operators.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}
	expectedKinds := []TokenType{
		TokenLPAREN,
		TokenRPAREN,
		TokenPLUS,
		TokenMINUS,
		TokenSTAR,
		TokenSLASH,
		TokenCARET,
		TokenMOD,
		TokenQUESTION,
		TokenCOLON,
		TokenSEMICOLON,
		TokenEQ,
		TokenEQEQ,
		TokenNE,
		TokenLE,
		TokenGE,
		TokenLT,
		TokenGT,
		TokenASSIGN,
	}
	assertTokenKinds(t, tokens, expectedKinds)

	mixed := NewLexer("if then else local foo _bar 12 3.5")
	mixedTokens, err := mixed.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}
	assertTokenKinds(t, mixedTokens, []TokenType{TokenIF, TokenTHEN, TokenELSE, TokenLOCAL, TokenIDENTIFIER, TokenIDENTIFIER, TokenINTEGER, TokenFLOAT})
	if mixedTokens[4].Lexeme != "foo" {
		t.Fatalf("mixed token 4 = %q", mixedTokens[4].Lexeme)
	}
	if mixedTokens[5].Lexeme != "_bar" {
		t.Fatalf("mixed token 5 = %q", mixedTokens[5].Lexeme)
	}

	stringsLexer := NewLexer("println \"hello\"\n-- comment\nprintln 'world'\n")
	stringTokens, err := stringsLexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}
	assertTokenKinds(t, stringTokens, []TokenType{TokenPRINTLN, TokenSTRING, TokenPRINTLN, TokenSTRING})
	if stringTokens[0].Line != 1 || stringTokens[2].Line != 3 {
		t.Fatalf("unexpected token lines: %d, %d", stringTokens[0].Line, stringTokens[2].Line)
	}
}

func makePositionedLexer(source string, current int, line int) *Lexer {
	lexer := NewLexer(source)
	lexer.tokenStart = 0
	lexer.current = current
	lexer.lineNumber = line
	return lexer
}

func assertTokensEqual(t *testing.T, got []Token, want []Token) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("token count = %d, want %d", len(got), len(want))
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("token %d = %+v, want %+v", index, got[index], want[index])
		}
	}
}

func assertTokenKinds(t *testing.T, got []Token, want []TokenType) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("token count = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index].Kind != want[index] {
			t.Fatalf("token kind %d = %v, want %v", index, got[index].Kind, want[index])
		}
	}
}
