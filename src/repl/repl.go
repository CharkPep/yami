package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/charkpep/yad/src/eval"
	"github.com/charkpep/yad/src/object"
	"github.com/charkpep/yad/src/parser"
	"io"
)

type Repl struct {
	lexerIn   io.Writer
	parser    *parser.Parser
	env       *object.Environment
	evaluator *eval.Evaluator
	in        io.Reader
	out       io.Writer
}

func New(in io.Reader, out io.Writer) *Repl {
	lexerIn := bytes.NewBuffer(make([]byte, 0))
	p := parser.NewParser(lexerIn)
	env := object.NewEnv()
	e := eval.NewEvaluator()
	return &Repl{
		lexerIn:   lexerIn,
		evaluator: e,
		env:       env,
		parser:    p,
		in:        in,
		out:       out,
	}
}

func (r Repl) Start() {
	s := bufio.NewScanner(r.in)
	for {
		fmt.Print(">> ")
		if !s.Scan() {
			return
		}

		r.lexerIn.Write(s.Bytes())
		root, err := r.parser.Parse()
		if err != nil {
			fmt.Println(err)
			continue
		}

		if len(r.parser.Errors) != 0 {
			fmt.Println(r.parser.Errors)
			continue
		}

		obj, err := r.evaluator.EvalWithEnv(root, r.env)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if obj != nil {
			fmt.Println(obj.Inspect())
		} else {
			fmt.Println("Nil")
		}

	}
}
