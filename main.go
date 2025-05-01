package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"

	"github.com/ChaosNyaruko/monkey/eval"
	"github.com/ChaosNyaruko/monkey/lexer"
	"github.com/ChaosNyaruko/monkey/object"
	"github.com/ChaosNyaruko/monkey/parser"
	"github.com/ChaosNyaruko/monkey/repl"
)

var (
	interactive = flag.Bool("i", false, "run the interpreter in interactive mode")
	filename    = flag.String("f", "", "the filename of source script to run")
	help        = flag.Bool("h", false, "show this help doc")
)

func main() {
	flag.Parse()
	if *help || len(flag.Args()) > 0 {
		flag.PrintDefaults()
		return
	}
	if *interactive || flag.NFlag() == 0 {
		user, err := user.Current()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Hello %s! This is the Monkey programming language!\nFeel to type in commands\n", user.Username)
		repl.Start(os.Stdin, os.Stdout)
		return
	}
	b, err := os.ReadFile(*filename)
	if err != nil {
		log.Fatalf("open file err: %v", err)
	}
	env := object.NewEnvironment(nil)
	srcCode := string(b)
	l := lexer.New(srcCode)
	p := parser.New(l)
	program := p.ParseProgram()
	if err = p.Error(); err != nil {
		io.WriteString(os.Stderr, err.Error())
	}

	// let reverse_sub = macro(a, b) {quote(unquote(b) - unquote(a))}
	if err := eval.DefineMacros(program, env); err != nil {
		fmt.Fprintf(os.Stderr, "define macros err: %v", err)
	}

	//  reverse_sub(1+2, 3+4) --> ((3+4)-(1+2))
	eval.ExpandMacros(program, env)

	// ((3+4)-(1+2)) -> 4
	_, err = eval.Eval(program, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "eval err: %v\n", err)
	}
	// error in interpreter
	if err != nil {
		os.Exit(1)
	}
	return
}
