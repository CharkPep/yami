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
		env:     make(map[string]Object),
	}
}

func (e Environment) Set(key string, val Object) {
	if _, ok := e.env[key]; ok {
		e.env[key] = val
		return
	}

	var parentEnv = e.rootEnv
	for parentEnv != nil {
		if _, ok := parentEnv.env[key]; ok {
			parentEnv.env[key] = val
			return
		}

		parentEnv = parentEnv.rootEnv
	}

	e.env[key] = val
}

func (e Environment) Get(key string) (Object, bool) {
	val, ok := e.env[key]
	if !ok && e.rootEnv != nil {
		return e.rootEnv.Get(key)
	}

	return val, ok
}
