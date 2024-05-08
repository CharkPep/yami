package object

type Environment struct {
	rootEnv *Environment
	env     map[string]Object
}

func NewEnv() *Environment {
	return &Environment{
		env: make(map[string]Object),
	}
}

func DeriveEnv(root *Environment) *Environment {
	return &Environment{
		rootEnv: root,
		env:     root.env,
	}
}

func (e Environment) Set(key string, val Object) {
	e.env[key] = val
}

func (e Environment) Get(key string) (Object, bool) {
	val, ok := e.env[key]
	if !ok && e.rootEnv != nil {
		return e.rootEnv.Get(key)
	}

	return val, ok
}
