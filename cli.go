package pinky

import (
	"fmt"
	"os"
	"strings"
)

type CLIIO struct {
	ReadTextFile func(path string) (string, error)
	WriteStdout  func(text string)
	WriteStderr  func(text string)
}

func defaultCLIIO() CLIIO {
	return CLIIO{
		ReadTextFile: func(path string) (string, error) {
			bytes, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		},
		WriteStdout: func(text string) {
			_, _ = os.Stdout.WriteString(text)
		},
		WriteStderr: func(text string) {
			_, _ = os.Stderr.WriteString(text)
		},
	}
}

func RunCLI(args []string, io CLIIO) int {
	if io.ReadTextFile == nil || io.WriteStdout == nil || io.WriteStderr == nil {
		io = defaultCLIIO()
	}

	if len(args) == 0 || containsArg(args, "--help") || containsArg(args, "-h") {
		io.WriteStderr("Usage: pinky-go <file.pinky> [--verbose]\n")
		if containsArg(args, "--help") || containsArg(args, "-h") {
			return 0
		}
		return 1
	}

	verbose := containsArg(args, "--verbose")
	filePath := firstPositionalArg(args)
	if filePath == "" {
		io.WriteStderr("Usage: pinky-go <file.pinky> [--verbose]\n")
		return 1
	}

	source, err := io.ReadTextFile(filePath)
	if err != nil {
		io.WriteStderr(fmt.Sprintf("Failed to read %s: %s\n", filePath, err.Error()))
		return 1
	}

	result := RunSource(source, verbose)
	if !result.Success {
		io.WriteStderr(fmt.Sprintf("%s error on line %d: %s\n", result.ErrorType, result.ErrorLine, result.ErrorMessage))
		return 1
	}

	io.WriteStdout(result.Output)
	if verbose {
		if len(result.Tokens) > 0 {
			io.WriteStderr("\nTokens\n" + strings.Join(result.Tokens, "\n") + "\n")
		}
		if result.AST != "" {
			io.WriteStderr("\nAST\n" + result.AST + "\n")
		}
	}

	return 0
}

func containsArg(args []string, target string) bool {
	for _, arg := range args {
		if arg == target {
			return true
		}
	}
	return false
}

func firstPositionalArg(args []string) string {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}
	return ""
}
