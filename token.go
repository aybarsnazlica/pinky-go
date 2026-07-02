package pinky

type Token struct {
	Kind   TokenType
	Lexeme string
	Line   int
}

func (t Token) DebugString() string {
	return "(" + tokenDebugName(t.Kind) + ", " + quoteDebugString(t.Lexeme) + ", " + stringify(t.Line) + ")"
}

func quoteDebugString(text string) string {
	result := "\""
	for _, ch := range text {
		switch ch {
		case '\\':
			result += "\\\\"
		case '"':
			result += "\\\""
		case '\n':
			result += "\\n"
		case '\t':
			result += "\\t"
		case '\r':
			result += "\\r"
		default:
			result += string(ch)
		}
	}
	return result + "\""
}
