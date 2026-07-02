package pinky

import "fmt"

type LexingError struct {
	Line        int
	MessageText string
}

func (e *LexingError) Error() string {
	return fmt.Sprintf("[Line %d]: %s", e.Line, e.MessageText)
}

type ParseError struct {
	Line        int
	MessageText string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("[Line %d]: %s", e.Line, e.MessageText)
}

type RuntimeError struct {
	Line        int
	MessageText string
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("[Line %d]: %s", e.Line, e.MessageText)
}
