package pinky

import "testing"

func TestEnvironmentLooksUpAndUpdatesVariablesThroughParents(t *testing.T) {
	global := NewEnvironment()
	global.SetLocal("x", NumberValue(1))

	child := global.NewEnv()
	if value, ok := child.GetVar("x"); !ok || value != NumberValue(1) {
		t.Fatalf("child.GetVar(x) = %+v, %v", value, ok)
	}

	child.SetVar("x", NumberValue(3))
	if value, ok := global.GetVar("x"); !ok || value != NumberValue(3) {
		t.Fatalf("global.GetVar(x) = %+v, %v", value, ok)
	}

	child.SetLocal("x", NumberValue(5))
	if value, ok := child.GetVar("x"); !ok || value != NumberValue(5) {
		t.Fatalf("child.GetVar(x) local = %+v, %v", value, ok)
	}
	if value, ok := global.GetVar("x"); !ok || value != NumberValue(3) {
		t.Fatalf("global.GetVar(x) after local = %+v, %v", value, ok)
	}
}

func TestEnvironmentStoresFunctionsWithDefiningScope(t *testing.T) {
	global := NewEnvironment()
	decl := &FunctionDecl{Name: "demo", Params: []*Param{}, BodyStmts: &Program{Statements: []Stmt{}, line: 1}, line: 1}
	global.SetFunc("demo", decl, nil)

	binding, ok := global.GetFunc("demo")
	if !ok {
		t.Fatal("GetFunc(demo) = false")
	}
	if binding.Declaration != decl {
		t.Fatal("unexpected declaration binding")
	}
	if binding.DefiningEnv != global {
		t.Fatal("unexpected defining env")
	}
	childBinding, ok := global.NewEnv().GetFunc("demo")
	if !ok || childBinding != binding {
		t.Fatal("child env did not inherit function binding")
	}
}
