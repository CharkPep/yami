package main

import (
	"github.com/charkpep/yami/src/eval"
	"github.com/charkpep/yami/src/parser"
	"github.com/charkpep/yami/src/repl"
	"io"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) == 1 {
		r := repl.New(os.Stdin, os.Stdout)
		r.Start()
	}

	f := os.Args[1]
	p, err := filepath.Abs(f)
	if err != nil {
		io.WriteString(os.Stdout, err.Error())
		os.Exit(1)
	}

	fd, err := os.OpenFile(p, os.O_RDONLY, 770)
	parser := parser.NewParser(fd)
	root, err := parser.Parse()
	if err != nil {
		io.WriteString(os.Stdout, err.Error())
		io.WriteString(os.Stdout, "\n")
		os.Exit(1)
	}

	if len(parser.Errors) != 0 {
		for _, err := range parser.Errors {
			io.WriteString(os.Stdout, err.Error())
			io.WriteString(os.Stdout, "\n")
		}
		os.Exit(1)
	}

	e := eval.NewEvaluator()
	if _, err := e.Eval(root); err != nil {
		io.WriteString(os.Stdout, err.Error())
		io.WriteString(os.Stdout, "\n")
		os.Exit(1)
	} else {
		//fmt.Println(val)
	}
}
