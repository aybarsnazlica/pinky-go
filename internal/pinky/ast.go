package pinky

import "strings"

type Node interface {
	String() string
	Line() int
}

type Expr interface {
	Node
	isExpr()
}

type Stmt interface {
	Node
	isStmt()
}

type IntegerLiteral struct {
	Value int
	line  int
}

func (n *IntegerLiteral) String() string { return "Integer[" + stringify(n.Value) + "]" }
func (n *IntegerLiteral) Line() int      { return n.line }
func (n *IntegerLiteral) isExpr()        {}

type FloatLiteral struct {
	Value float64
	line  int
}

func (n *FloatLiteral) String() string { return "Float[" + stringify(n.Value) + "]" }
func (n *FloatLiteral) Line() int      { return n.line }
func (n *FloatLiteral) isExpr()        {}

type BoolLiteral struct {
	Value bool
	line  int
}

func (n *BoolLiteral) String() string { return "Bool[" + stringify(n.Value) + "]" }
func (n *BoolLiteral) Line() int      { return n.line }
func (n *BoolLiteral) isExpr()        {}

type StringLiteral struct {
	Value string
	line  int
}

func (n *StringLiteral) String() string { return "String[" + n.Value + "]" }
func (n *StringLiteral) Line() int      { return n.line }
func (n *StringLiteral) isExpr()        {}

type UnaryExpr struct {
	Op      Token
	Operand Expr
	line    int
}

func (n *UnaryExpr) String() string {
	return "UnOp(\"" + n.Op.Lexeme + "\", " + n.Operand.String() + ")"
}
func (n *UnaryExpr) Line() int { return n.line }
func (n *UnaryExpr) isExpr()   {}

type BinaryExpr struct {
	Op    Token
	Left  Expr
	Right Expr
	line  int
}

func (n *BinaryExpr) String() string {
	return "BinOp(\"" + n.Op.Lexeme + "\", " + n.Left.String() + ", " + n.Right.String() + ")"
}

func (n *BinaryExpr) Line() int { return n.line }
func (n *BinaryExpr) isExpr()   {}

type LogicalExpr struct {
	Op    Token
	Left  Expr
	Right Expr
	line  int
}

func (n *LogicalExpr) String() string {
	return "LogicalOp(\"" + n.Op.Lexeme + "\", " + n.Left.String() + ", " + n.Right.String() + ")"
}

func (n *LogicalExpr) Line() int { return n.line }
func (n *LogicalExpr) isExpr()   {}

type GroupingExpr struct {
	Value Expr
	line  int
}

func (n *GroupingExpr) String() string { return "Grouping(" + n.Value.String() + ")" }
func (n *GroupingExpr) Line() int      { return n.line }
func (n *GroupingExpr) isExpr()        {}

type IdentifierExpr struct {
	Name string
	line int
}

func (n *IdentifierExpr) String() string { return "Identifier[\"" + n.Name + "\"]" }
func (n *IdentifierExpr) Line() int      { return n.line }
func (n *IdentifierExpr) isExpr()        {}

type Program struct {
	Statements []Stmt
	line       int
}

func (n *Program) String() string { return "Stmts([" + joinNodes(n.Statements) + "])" }
func (n *Program) Line() int      { return n.line }

type PrintStmt struct {
	Value Expr
	End   string
	line  int
}

func (n *PrintStmt) String() string {
	renderedEnd := n.End
	if renderedEnd == "\n" {
		renderedEnd = "\\n"
	}
	return "PrintStmt(" + n.Value.String() + ", end=\"" + renderedEnd + "\")"
}

func (n *PrintStmt) Line() int { return n.line }
func (n *PrintStmt) isStmt()   {}

type IfStmt struct {
	Test      Expr
	ThenStmts *Program
	ElseStmts *Program
	line      int
}

func (n *IfStmt) String() string {
	elseStmts := "None"
	if n.ElseStmts != nil {
		elseStmts = n.ElseStmts.String()
	}
	return "IfStmt(" + n.Test.String() + ", then:" + n.ThenStmts.String() + ", else:" + elseStmts + ")"
}

func (n *IfStmt) Line() int { return n.line }
func (n *IfStmt) isStmt()   {}

type WhileStmt struct {
	Test      Expr
	BodyStmts *Program
	line      int
}

func (n *WhileStmt) String() string {
	return "WhileStmt(" + n.Test.String() + ", " + n.BodyStmts.String() + ")"
}
func (n *WhileStmt) Line() int { return n.line }
func (n *WhileStmt) isStmt()   {}

type AssignmentStmt struct {
	Left  Expr
	Right Expr
	line  int
}

func (n *AssignmentStmt) String() string {
	return "Assignment(" + n.Left.String() + ", " + n.Right.String() + ")"
}
func (n *AssignmentStmt) Line() int { return n.line }
func (n *AssignmentStmt) isStmt()   {}

type LocalAssignmentStmt struct {
	Left  Expr
	Right Expr
	line  int
}

func (n *LocalAssignmentStmt) String() string {
	return "LocalAssignment(" + n.Left.String() + ", " + n.Right.String() + ")"
}

func (n *LocalAssignmentStmt) Line() int { return n.line }
func (n *LocalAssignmentStmt) isStmt()   {}

type ForStmt struct {
	Ident     *IdentifierExpr
	Start     Expr
	End       Expr
	Step      Expr
	BodyStmts *Program
	line      int
}

func (n *ForStmt) String() string {
	step := "None"
	if n.Step != nil {
		step = n.Step.String()
	}
	return "ForStmt(" + n.Ident.String() + ", " + n.Start.String() + ", " + n.End.String() + ", " + step + ", " + n.BodyStmts.String() + ")"
}

func (n *ForStmt) Line() int { return n.line }
func (n *ForStmt) isStmt()   {}

type Param struct {
	Name string
	line int
}

func (n *Param) String() string { return "Param[\"" + n.Name + "\"]" }
func (n *Param) Line() int      { return n.line }

type FunctionDecl struct {
	Name      string
	Params    []*Param
	BodyStmts *Program
	line      int
}

func (n *FunctionDecl) String() string {
	return "FuncDecl(\"" + n.Name + "\", [" + joinNodes(n.Params) + "], " + n.BodyStmts.String() + ")"
}

func (n *FunctionDecl) Line() int { return n.line }
func (n *FunctionDecl) isStmt()   {}

type FunctionCallExpr struct {
	Name string
	Args []Expr
	line int
}

func (n *FunctionCallExpr) String() string {
	return "FuncCall(\"" + n.Name + "\", [" + joinNodes(n.Args) + "])"
}
func (n *FunctionCallExpr) Line() int { return n.line }
func (n *FunctionCallExpr) isExpr()   {}

type FunctionCallStmt struct {
	ExprValue *FunctionCallExpr
}

func (n *FunctionCallStmt) String() string { return "FuncCallStmt(" + n.ExprValue.String() + ")" }
func (n *FunctionCallStmt) Line() int      { return n.ExprValue.Line() }
func (n *FunctionCallStmt) isStmt()        {}

type ReturnStmt struct {
	Value Expr
	line  int
}

func (n *ReturnStmt) String() string { return "RetStmt[" + n.Value.String() + "]" }
func (n *ReturnStmt) Line() int      { return n.line }
func (n *ReturnStmt) isStmt()        {}

func joinNodes[T interface{ String() string }](nodes []T) string {
	parts := make([]string, 0, len(nodes))
	for _, node := range nodes {
		parts = append(parts, node.String())
	}
	return strings.Join(parts, ", ")
}
