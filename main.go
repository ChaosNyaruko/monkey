package main

import (
	"bufio"
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
	f, err := os.Open(*filename)
	if err != nil {
		log.Fatalf("open file err: %v", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	env := object.NewEnvironment(nil)
	for scanner.Scan() {
		l := lexer.New(scanner.Text())
		p := parser.New(l)
		program := p.ParseProgram()
		if err = p.Error(); err != nil {
			io.WriteString(os.Stderr, err.Error())
			continue
		}
		ob, err := eval.Eval(program, env)
		if err != nil {
			fmt.Fprintf(os.Stderr, "eval err: %v", err)
			break
		}
		if ob == nil { // EOF reached
			continue
		}
		fmt.Fprintf(os.Stdout, "%s\n", ob.Inspect())
	}
	if err = scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading file error: %v\n", err)
	}
	// error in interpreter
	if err != nil {
		os.Exit(1)
	}
	return
}
