package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/charkpep/yad/src/eval"
	"github.com/charkpep/yad/src/lexer"
	"github.com/charkpep/yad/src/parser"
	"io"
)

type Repl struct {
	lexer     lexer.ILexer
	lexerIn   io.Writer
	parser    *parser.Parser
	evaluator *eval.Evaluator
	in        io.Reader
	out       io.Writer
}

func New(in io.Reader, out io.Writer) *Repl {
	lexerIn := bytes.NewBuffer(make([]byte, 0))
	lex := lexer.New(lexerIn)
	p := parser.NewParserFromLexer(lex)
	e := eval.NewEvaluator()
	return &Repl{
		lexer:     lex,
		lexerIn:   lexerIn,
		evaluator: e,
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
		for r.lexer.CurToken().Token == lexer.EOF && len(s.Bytes()) != 0 {
			r.lexer.Advance()
		}

		root, err := r.parser.Parse()
		if err != nil {
			fmt.Println(err)
			continue
		}

		if len(r.parser.Errors) != 0 {
			fmt.Println(r.parser.Errors)
			continue
		}

		obj, err := r.evaluator.Eval(root)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println(obj.Inspect())
	}
}
