package pinky

type RuntimeType string

const (
	RuntimeTypeNumber RuntimeType = "NUMBER"
	RuntimeTypeString RuntimeType = "STRING"
	RuntimeTypeBool   RuntimeType = "BOOL"
)

type RuntimeValue struct {
	Type   RuntimeType
	Number float64
	String string
	Bool   bool
}

func NumberValue(value float64) RuntimeValue {
	return RuntimeValue{Type: RuntimeTypeNumber, Number: value}
}

func StringValue(value string) RuntimeValue {
	return RuntimeValue{Type: RuntimeTypeString, String: value}
}

func BoolValue(value bool) RuntimeValue {
	return RuntimeValue{Type: RuntimeTypeBool, Bool: value}
}

type FunctionBinding struct {
	Declaration *FunctionDecl
	DefiningEnv *Environment
}

func runtimeValueToString(value RuntimeValue) string {
	switch value.Type {
	case RuntimeTypeNumber:
		return stringify(value.Number)
	case RuntimeTypeString:
		return stringify(value.String)
	case RuntimeTypeBool:
		return stringify(value.Bool)
	default:
		return "<nil>"
	}
}

type Environment struct {
	vars      map[string]RuntimeValue
	funcs     map[string]*FunctionBinding
	parentEnv *Environment
}

func NewEnvironment() *Environment {
	return &Environment{vars: map[string]RuntimeValue{}, funcs: map[string]*FunctionBinding{}}
}

func (e *Environment) GetVar(name string) (RuntimeValue, bool) {
	for env := e; env != nil; env = env.parentEnv {
		if value, ok := env.vars[name]; ok {
			return value, true
		}
	}
	return RuntimeValue{}, false
}

func (e *Environment) SetVar(name string, value RuntimeValue) {
	for env := e; env != nil; env = env.parentEnv {
		if _, ok := env.vars[name]; ok {
			env.vars[name] = value
			return
		}
	}
	e.vars[name] = value
}

func (e *Environment) SetLocal(name string, value RuntimeValue) {
	e.vars[name] = value
}

func (e *Environment) GetFunc(name string) (*FunctionBinding, bool) {
	for env := e; env != nil; env = env.parentEnv {
		if binding, ok := env.funcs[name]; ok {
			return binding, true
		}
	}
	return nil, false
}

func (e *Environment) SetFunc(name string, declaration *FunctionDecl, definingEnv *Environment) {
	if definingEnv == nil {
		definingEnv = e
	}
	e.funcs[name] = &FunctionBinding{Declaration: declaration, DefiningEnv: definingEnv}
}

func (e *Environment) NewEnv() *Environment {
	return &Environment{vars: map[string]RuntimeValue{}, funcs: map[string]*FunctionBinding{}, parentEnv: e}
}

func (e *Environment) Parent() *Environment {
	return e.parentEnv
}
