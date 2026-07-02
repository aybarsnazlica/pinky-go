package pinky

type RunResult struct {
	Success      bool     `json:"success"`
	Source       string   `json:"source"`
	Tokens       []string `json:"tokens"`
	AST          string   `json:"ast"`
	Output       string   `json:"output"`
	ErrorType    string   `json:"errorType"`
	ErrorMessage string   `json:"errorMessage"`
	ErrorLine    int      `json:"errorLine"`
}

func RunSource(source string, includeDebug bool) RunResult {
	lexer := NewLexer(source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		switch typed := err.(type) {
		case *LexingError:
			return failure(source, "lex", typed.MessageText, typed.Line)
		default:
			return failure(source, "internal", err.Error(), 0)
		}
	}

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		switch typed := err.(type) {
		case *ParseError:
			return failure(source, "parse", typed.MessageText, typed.Line)
		default:
			return failure(source, "internal", err.Error(), 0)
		}
	}

	output := ""
	interpreter := NewInterpreter(func(text string) {
		output += text
	})
	if _, err := interpreter.InterpretAST(ast); err != nil {
		switch typed := err.(type) {
		case *RuntimeError:
			return failure(source, "runtime", typed.MessageText, typed.Line)
		default:
			return failure(source, "internal", err.Error(), 0)
		}
	}

	result := RunResult{
		Success: true,
		Source:  source,
		Output:  output,
	}
	if includeDebug {
		result.Tokens = make([]string, 0, len(tokens))
		for _, token := range tokens {
			result.Tokens = append(result.Tokens, token.DebugString())
		}
		result.AST = PrintPrettyAST(ast.String())
	}
	return result
}

func failure(source string, errorType string, errorMessage string, errorLine int) RunResult {
	return RunResult{
		Success:      false,
		Source:       source,
		Tokens:       []string{},
		AST:          "",
		Output:       "",
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
		ErrorLine:    errorLine,
	}
}
