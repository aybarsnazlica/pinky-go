package pinky

import (
	"math"
	"strconv"
	"strings"
)

func PrintPrettyAST(astText string) string {
	var result strings.Builder
	indentation := 0
	newline := false

	for _, ch := range astText {
		switch ch {
		case '(':
			result.WriteRune(ch)
			result.WriteByte('\n')
			indentation += 2
			newline = true
		case ')':
			if !newline {
				result.WriteByte('\n')
			}
			indentation -= 2
			if indentation < 0 {
				indentation = 0
			}
			result.WriteString(strings.Repeat(" ", indentation))
			result.WriteRune(ch)
			newline = true
		default:
			if newline {
				result.WriteString(strings.Repeat(" ", indentation))
			}
			result.WriteRune(ch)
			newline = false
		}
	}

	return result.String()
}

func stringify(value any) string {
	switch v := value.(type) {
	case nil:
		return "<nil>"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(v)
	case float64:
		if !math.IsNaN(v) && !math.IsInf(v, 0) && math.Trunc(v) == v {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	default:
		return strconv.FormatBool(false)
	}
}

func quote(text string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return "'" + replacer.Replace(text) + "'"
}
