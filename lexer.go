package pinky

import "unicode"

var keywords = map[string]TokenType{
	"if":      TokenIF,
	"else":    TokenELSE,
	"then":    TokenTHEN,
	"true":    TokenTRUE,
	"false":   TokenFALSE,
	"and":     TokenAND,
	"or":      TokenOR,
	"local":   TokenLOCAL,
	"while":   TokenWHILE,
	"do":      TokenDO,
	"for":     TokenFOR,
	"func":    TokenFUNC,
	"null":    TokenNULL,
	"end":     TokenEND,
	"print":   TokenPRINT,
	"println": TokenPRINTLN,
	"ret":     TokenRET,
}

type Lexer struct {
	Source     string
	tokenStart int
	current    int
	lineNumber int
	tokens     []Token
}

func NewLexer(source string) *Lexer {
	return &Lexer{Source: source, lineNumber: 1}
}

func (l *Lexer) advance() byte {
	ch := l.Source[l.current]
	l.current++
	return ch
}

func (l *Lexer) peek() byte {
	if l.current >= len(l.Source) {
		return '\x00'
	}
	return l.Source[l.current]
}

func (l *Lexer) lookahead(n int) byte {
	index := l.current + n
	if index < 0 || index >= len(l.Source) {
		return '\x00'
	}
	return l.Source[index]
}

func (l *Lexer) match(expected byte) bool {
	if l.current >= len(l.Source) {
		return false
	}
	if l.Source[l.current] != expected {
		return false
	}
	l.current++
	return true
}

func (l *Lexer) handleNumber() {
	for isDigit(l.peek()) {
		l.advance()
	}

	if l.peek() == '.' && isDigit(l.lookahead(1)) {
		l.advance()
		for isDigit(l.peek()) {
			l.advance()
		}
		l.addToken(TokenFLOAT)
		return
	}

	l.addToken(TokenINTEGER)
}

func (l *Lexer) handleIdentifier() {
	for isAlphaNumeric(l.peek()) || l.peek() == '_' {
		l.advance()
	}

	text := l.Source[l.tokenStart:l.current]
	if keyword, ok := keywords[text]; ok {
		l.addToken(keyword)
		return
	}
	l.addToken(TokenIDENTIFIER)
}

func (l *Lexer) addToken(kind TokenType) {
	l.tokens = append(l.tokens, Token{Kind: kind, Lexeme: l.Source[l.tokenStart:l.current], Line: l.lineNumber})
}

func (l *Lexer) handleString(startQuote byte) error {
	for l.peek() != startQuote && l.current < len(l.Source) {
		l.advance()
	}

	if l.current >= len(l.Source) {
		return &LexingError{Line: l.lineNumber, MessageText: "Unterminated string."}
	}

	l.advance()
	l.addToken(TokenSTRING)
	return nil
}

func (l *Lexer) Tokenize() ([]Token, error) {
	l.tokenStart = 0
	l.current = 0
	l.lineNumber = 1
	l.tokens = l.tokens[:0]

	for l.current < len(l.Source) {
		l.tokenStart = l.current
		ch := l.advance()

		switch ch {
		case '\n':
			l.lineNumber++
		case ' ', '\t', '\r':
		case '(':
			l.addToken(TokenLPAREN)
		case ')':
			l.addToken(TokenRPAREN)
		case '{':
			l.addToken(TokenLCURLY)
		case '}':
			l.addToken(TokenRCURLY)
		case '[':
			l.addToken(TokenLSQUAR)
		case ']':
			l.addToken(TokenRSQUAR)
		case '.':
			l.addToken(TokenDOT)
		case ',':
			l.addToken(TokenCOMMA)
		case '+':
			l.addToken(TokenPLUS)
		case '*':
			l.addToken(TokenSTAR)
		case '^':
			l.addToken(TokenCARET)
		case '/':
			l.addToken(TokenSLASH)
		case ';':
			l.addToken(TokenSEMICOLON)
		case '?':
			l.addToken(TokenQUESTION)
		case '%':
			l.addToken(TokenMOD)
		case '-':
			if l.match('-') {
				for l.peek() != '\n' && l.current < len(l.Source) {
					l.advance()
				}
			} else {
				l.addToken(TokenMINUS)
			}
		case '=':
			if l.match('=') {
				l.addToken(TokenEQEQ)
			} else {
				l.addToken(TokenEQ)
			}
		case '~':
			if l.match('=') {
				l.addToken(TokenNE)
			} else {
				l.addToken(TokenNOT)
			}
		case '<':
			if l.match('=') {
				l.addToken(TokenLE)
			} else {
				l.addToken(TokenLT)
			}
		case '>':
			if l.match('=') {
				l.addToken(TokenGE)
			} else {
				l.addToken(TokenGT)
			}
		case ':':
			if l.match('=') {
				l.addToken(TokenASSIGN)
			} else {
				l.addToken(TokenCOLON)
			}
		case '"', '\'':
			if err := l.handleString(ch); err != nil {
				return nil, err
			}
		default:
			if isDigit(ch) {
				l.handleNumber()
			} else if isAlpha(ch) || ch == '_' {
				l.handleIdentifier()
			} else {
				return nil, &LexingError{Line: l.lineNumber, MessageText: "Error at " + quoteChar(ch) + ": Unexpected character."}
			}
		}
	}

	return append([]Token(nil), l.tokens...), nil
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isAlpha(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isAlphaNumeric(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

func quoteChar(ch byte) string {
	switch ch {
	case '\x00':
		return "'\\0'"
	case '\n':
		return "'\\n'"
	case '\t':
		return "'\\t'"
	case '\r':
		return "'\\r'"
	case '\'':
		return "'\\''"
	case '\\':
		return "'\\\\'"
	default:
		return quote(string(ch))
	}
}
