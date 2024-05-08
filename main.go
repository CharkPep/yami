package main

import (
	"github.com/charkpep/yad/src/repl"
	"os"
)

func main() {
	r := repl.New(os.Stdin, os.Stdout)
	r.Start()
}
