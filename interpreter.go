package pinky

import (
	"fmt"
	"math"
	"os"
)

type returnSignal struct {
	value RuntimeValue
}

func (s *returnSignal) Error() string { return "return" }

type Interpreter struct {
	out func(string)
}

func NewInterpreter(out func(string)) *Interpreter {
	if out == nil {
		out = func(text string) {
			_, _ = os.Stdout.WriteString(text)
		}
	}
	return &Interpreter{out: out}
}

func (i *Interpreter) Interpret(node Node, env *Environment) (*RuntimeValue, error) {
	if env == nil {
		env = NewEnvironment()
	}

	switch n := node.(type) {
	case *FunctionCallExpr:
		return i.invokeFunction(n, env)
	case *Program:
		return nil, i.Execute(n, env)
	case *Param:
		return nil, &RuntimeError{Line: n.Line(), MessageText: "Unsupported AST node."}
	}

	if stmt, ok := node.(Stmt); ok {
		return nil, i.Execute(stmt, env)
	}

	expr, ok := node.(Expr)
	if !ok {
		return nil, &RuntimeError{Line: 0, MessageText: "Unsupported AST node."}
	}
	value, err := i.Evaluate(expr, env)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (i *Interpreter) Evaluate(expr Expr, env *Environment) (RuntimeValue, error) {
	switch n := expr.(type) {
	case *IntegerLiteral:
		return NumberValue(float64(n.Value)), nil
	case *FloatLiteral:
		return NumberValue(n.Value), nil
	case *StringLiteral:
		return StringValue(n.Value), nil
	case *BoolLiteral:
		return BoolValue(n.Value), nil
	case *GroupingExpr:
		return i.Evaluate(n.Value, env)
	case *IdentifierExpr:
		slot, ok := env.GetVar(n.Name)
		if !ok {
			return RuntimeValue{}, &RuntimeError{Line: n.Line(), MessageText: "Undeclared identifier " + quote(n.Name)}
		}
		return slot, nil
	case *BinaryExpr:
		left, err := i.Evaluate(n.Left, env)
		if err != nil {
			return RuntimeValue{}, err
		}
		right, err := i.Evaluate(n.Right, env)
		if err != nil {
			return RuntimeValue{}, err
		}
		return interpretBinary(n, left, right)
	case *UnaryExpr:
		operand, err := i.Evaluate(n.Operand, env)
		if err != nil {
			return RuntimeValue{}, err
		}
		return interpretUnary(n, operand)
	case *LogicalExpr:
		left, err := i.Evaluate(n.Left, env)
		if err != nil {
			return RuntimeValue{}, err
		}
		if n.Op.Kind == TokenOR {
			if isTruthy(left) {
				return left, nil
			}
		} else if n.Op.Kind == TokenAND {
			if !isTruthy(left) {
				return left, nil
			}
		}
		return i.Evaluate(n.Right, env)
	case *FunctionCallExpr:
		result, err := i.invokeFunction(n, env)
		if err != nil {
			return RuntimeValue{}, err
		}
		if result != nil {
			return *result, nil
		}
		return RuntimeValue{}, &RuntimeError{Line: n.Line(), MessageText: "Function " + quote(n.Name) + " did not return a value."}
	default:
		return RuntimeValue{}, &RuntimeError{Line: 0, MessageText: "Unsupported expression."}
	}
}

func (i *Interpreter) Execute(program Node, env *Environment) error {
	if block, ok := program.(*Program); ok {
		for _, stmt := range block.Statements {
			if err := i.Execute(stmt, env); err != nil {
				return err
			}
		}
		return nil
	}

	switch n := program.(type) {
	case *AssignmentStmt:
		identifier, ok := n.Left.(*IdentifierExpr)
		if !ok {
			return &RuntimeError{Line: n.Line(), MessageText: "Assignment target must be an identifier."}
		}
		value, err := i.Evaluate(n.Right, env)
		if err != nil {
			return err
		}
		env.SetVar(identifier.Name, value)
		return nil
	case *LocalAssignmentStmt:
		identifier, ok := n.Left.(*IdentifierExpr)
		if !ok {
			return &RuntimeError{Line: n.Line(), MessageText: "Assignment target must be an identifier."}
		}
		value, err := i.Evaluate(n.Right, env)
		if err != nil {
			return err
		}
		env.SetLocal(identifier.Name, value)
		return nil
	case *PrintStmt:
		value, err := i.Evaluate(n.Value, env)
		if err != nil {
			return err
		}
		i.out(decodeEscapes(runtimeValueToString(value)) + n.End)
		return nil
	case *IfStmt:
		test, err := i.Evaluate(n.Test, env)
		if err != nil {
			return err
		}
		if test.Type != RuntimeTypeBool {
			return &RuntimeError{Line: n.Line(), MessageText: "Condition test is not a boolean expression."}
		}
		if asBool(test) {
			return i.Execute(n.ThenStmts, env.NewEnv())
		}
		if n.ElseStmts != nil {
			return i.Execute(n.ElseStmts, env.NewEnv())
		}
		return nil
	case *WhileStmt:
		blockEnv := env.NewEnv()
		for {
			test, err := i.Evaluate(n.Test, env)
			if err != nil {
				return err
			}
			if test.Type != RuntimeTypeBool {
				return &RuntimeError{Line: n.Line(), MessageText: "While test is not a boolean expression."}
			}
			if !asBool(test) {
				break
			}
			if err := i.Execute(n.BodyStmts, blockEnv); err != nil {
				return err
			}
		}
		return nil
	case *ForStmt:
		start, err := i.Evaluate(n.Start, env)
		if err != nil {
			return err
		}
		end, err := i.Evaluate(n.End, env)
		if err != nil {
			return err
		}
		if start.Type != RuntimeTypeNumber || end.Type != RuntimeTypeNumber {
			return &RuntimeError{Line: n.Line(), MessageText: "For range bounds must be numbers."}
		}

		current := asNumber(start)
		limit := asNumber(end)
		blockEnv := env.NewEnv()
		step := 1.0
		if current >= limit {
			step = -1.0
		}

		if n.Step != nil {
			stepValue, err := i.Evaluate(n.Step, env)
			if err != nil {
				return err
			}
			if stepValue.Type != RuntimeTypeNumber {
				return &RuntimeError{Line: n.Line(), MessageText: "For step must be a number."}
			}
			step = asNumber(stepValue)
		}

		if current < limit {
			for current <= limit {
				env.SetVar(n.Ident.Name, NumberValue(current))
				if err := i.Execute(n.BodyStmts, blockEnv); err != nil {
					return err
				}
				current += step
			}
			return nil
		}

		for current >= limit {
			env.SetVar(n.Ident.Name, NumberValue(current))
			if err := i.Execute(n.BodyStmts, blockEnv); err != nil {
				return err
			}
			current += step
		}
		return nil
	case *FunctionDecl:
		env.SetFunc(n.Name, n, env)
		return nil
	case *FunctionCallStmt:
		_, err := i.invokeFunction(n.ExprValue, env)
		return err
	case *ReturnStmt:
		value, err := i.Evaluate(n.Value, env)
		if err != nil {
			return err
		}
		return &returnSignal{value: value}
	default:
		return nil
	}
}

func (i *Interpreter) InterpretAST(node Node) (*Environment, error) {
	env := NewEnvironment()
	if _, err := i.Interpret(node, env); err != nil {
		return nil, err
	}
	return env, nil
}

func (i *Interpreter) invokeFunction(functionCallExpr *FunctionCallExpr, env *Environment) (*RuntimeValue, error) {
	functionBinding, ok := env.GetFunc(functionCallExpr.Name)
	if !ok {
		return nil, &RuntimeError{Line: functionCallExpr.Line(), MessageText: "Function " + quote(functionCallExpr.Name) + " not declared."}
	}
	if len(functionCallExpr.Args) != len(functionBinding.Declaration.Params) {
		return nil, &RuntimeError{Line: functionCallExpr.Line(), MessageText: fmt.Sprintf("Function %s expected %d params but %d args were passed.", quote(functionBinding.Declaration.Name), len(functionBinding.Declaration.Params), len(functionCallExpr.Args))}
	}

	args := make([]RuntimeValue, 0, len(functionCallExpr.Args))
	for _, arg := range functionCallExpr.Args {
		value, err := i.Evaluate(arg, env)
		if err != nil {
			return nil, err
		}
		args = append(args, value)
	}

	newFuncEnv := functionBinding.DefiningEnv.NewEnv()
	for index, param := range functionBinding.Declaration.Params {
		newFuncEnv.SetLocal(param.Name, args[index])
	}

	if err := i.Execute(functionBinding.Declaration.BodyStmts, newFuncEnv); err != nil {
		if signal, ok := err.(*returnSignal); ok {
			return &signal.value, nil
		}
		return nil, err
	}

	return nil, nil
}

func interpretBinary(binaryExpr *BinaryExpr, left RuntimeValue, right RuntimeValue) (RuntimeValue, error) {
	switch binaryExpr.Op.Kind {
	case TokenPLUS:
		if left.Type == RuntimeTypeNumber && right.Type == RuntimeTypeNumber {
			return NumberValue(asNumber(left) + asNumber(right)), nil
		}
		if left.Type == RuntimeTypeString || right.Type == RuntimeTypeString {
			return StringValue(runtimeValueToString(left) + runtimeValueToString(right)), nil
		}
		return RuntimeValue{}, unsupportedBinary(binaryExpr, left, right)
	case TokenMINUS:
		return numericBinary(binaryExpr, left, right, func(a, b float64) float64 { return a - b })
	case TokenSTAR:
		return numericBinary(binaryExpr, left, right, func(a, b float64) float64 { return a * b })
	case TokenSLASH:
		if right.Type == RuntimeTypeNumber && asNumber(right) == 0 {
			return RuntimeValue{}, &RuntimeError{Line: binaryExpr.Line(), MessageText: "Division by zero."}
		}
		return numericBinary(binaryExpr, left, right, func(a, b float64) float64 { return a / b })
	case TokenMOD:
		return numericBinary(binaryExpr, left, right, signedModulo)
	case TokenCARET:
		return numericBinary(binaryExpr, left, right, math.Pow)
	case TokenGT:
		return compare(binaryExpr, left, right, "GT")
	case TokenGE:
		return compare(binaryExpr, left, right, "GE")
	case TokenLT:
		return compare(binaryExpr, left, right, "LT")
	case TokenLE:
		return compare(binaryExpr, left, right, "LE")
	case TokenEQEQ:
		if left.Type == right.Type {
			return BoolValue(valuesEqual(left, right)), nil
		}
		return RuntimeValue{}, unsupportedBinary(binaryExpr, left, right)
	case TokenNE:
		if left.Type == right.Type {
			return BoolValue(!valuesEqual(left, right)), nil
		}
		return RuntimeValue{}, unsupportedBinary(binaryExpr, left, right)
	default:
		return RuntimeValue{}, unsupportedBinary(binaryExpr, left, right)
	}
}

func interpretUnary(unaryExpr *UnaryExpr, operand RuntimeValue) (RuntimeValue, error) {
	switch unaryExpr.Op.Kind {
	case TokenMINUS:
		if operand.Type == RuntimeTypeNumber {
			return NumberValue(-asNumber(operand)), nil
		}
		return RuntimeValue{}, unsupportedUnary(unaryExpr, operand)
	case TokenPLUS:
		if operand.Type == RuntimeTypeNumber {
			return NumberValue(asNumber(operand)), nil
		}
		return RuntimeValue{}, unsupportedUnary(unaryExpr, operand)
	case TokenNOT:
		if operand.Type == RuntimeTypeBool {
			return BoolValue(!asBool(operand)), nil
		}
		return RuntimeValue{}, unsupportedUnary(unaryExpr, operand)
	default:
		return RuntimeValue{}, unsupportedUnary(unaryExpr, operand)
	}
}

func numericBinary(binaryExpr *BinaryExpr, left RuntimeValue, right RuntimeValue, operator func(a, b float64) float64) (RuntimeValue, error) {
	if left.Type == RuntimeTypeNumber && right.Type == RuntimeTypeNumber {
		return NumberValue(operator(asNumber(left), asNumber(right))), nil
	}
	return RuntimeValue{}, unsupportedBinary(binaryExpr, left, right)
}

func compare(binaryExpr *BinaryExpr, left RuntimeValue, right RuntimeValue, operator string) (RuntimeValue, error) {
	if left.Type == RuntimeTypeNumber && right.Type == RuntimeTypeNumber {
		leftNumber := asNumber(left)
		rightNumber := asNumber(right)
		switch operator {
		case "GT":
			return BoolValue(leftNumber > rightNumber), nil
		case "GE":
			return BoolValue(leftNumber >= rightNumber), nil
		case "LT":
			return BoolValue(leftNumber < rightNumber), nil
		case "LE":
			return BoolValue(leftNumber <= rightNumber), nil
		}
	}

	if left.Type == RuntimeTypeString && right.Type == RuntimeTypeString {
		switch operator {
		case "GT":
			return BoolValue(left.String > right.String), nil
		case "GE":
			return BoolValue(left.String >= right.String), nil
		case "LT":
			return BoolValue(left.String < right.String), nil
		case "LE":
			return BoolValue(left.String <= right.String), nil
		}
	}

	return RuntimeValue{}, unsupportedBinary(binaryExpr, left, right)
}

func unsupportedBinary(binaryExpr *BinaryExpr, left RuntimeValue, right RuntimeValue) error {
	return &RuntimeError{Line: binaryExpr.Op.Line, MessageText: "Unsupported operator " + quote(binaryExpr.Op.Lexeme) + " between " + typeName(left) + " and " + typeName(right) + "."}
}

func unsupportedUnary(unaryExpr *UnaryExpr, operand RuntimeValue) error {
	return &RuntimeError{Line: unaryExpr.Op.Line, MessageText: "Unsupported operator " + quote(unaryExpr.Op.Lexeme) + " with " + typeName(operand) + "."}
}

func asNumber(value RuntimeValue) float64 {
	return value.Number
}

func asBool(value RuntimeValue) bool {
	return value.Bool
}

func isTruthy(value RuntimeValue) bool {
	switch value.Type {
	case RuntimeTypeNumber:
		return asNumber(value) != 0
	case RuntimeTypeString:
		return len(value.String) > 0
	case RuntimeTypeBool:
		return asBool(value)
	default:
		return false
	}
}

func typeName(value RuntimeValue) string {
	switch value.Type {
	case RuntimeTypeNumber:
		return "TYPE_NUMBER"
	case RuntimeTypeString:
		return "TYPE_STRING"
	case RuntimeTypeBool:
		return "TYPE_BOOL"
	default:
		return "TYPE_UNKNOWN"
	}
}

func valuesEqual(left RuntimeValue, right RuntimeValue) bool {
	switch left.Type {
	case RuntimeTypeNumber:
		return left.Number == right.Number
	case RuntimeTypeString:
		return left.String == right.String
	case RuntimeTypeBool:
		return left.Bool == right.Bool
	default:
		return false
	}
}

func decodeEscapes(input string) string {
	output := ""
	for index := 0; index < len(input); index++ {
		ch := input[index]
		if ch != '\\' || index+1 >= len(input) {
			output += string(ch)
			continue
		}

		index++
		switch input[index] {
		case 'n':
			output += "\n"
		case 't':
			output += "\t"
		case 'r':
			output += "\r"
		case '\\':
			output += "\\"
		case '"':
			output += `"`
		case '\'':
			output += "'"
		default:
			output += "\\" + string(input[index])
		}
	}
	return output
}

func signedModulo(left float64, right float64) float64 {
	result := math.Mod(left, right)
	if result != 0 && ((result < 0) != (right < 0)) {
		result += right
	}
	return result
}
