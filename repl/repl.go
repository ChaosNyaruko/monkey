package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/ChaosNyaruko/monkey/eval"
	"github.com/ChaosNyaruko/monkey/lexer"
	"github.com/ChaosNyaruko/monkey/parser"
	"github.com/ChaosNyaruko/monkey/token"
)

const PROMPT = ">>"

const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

// read-evaluate-print loop
func Start(in io.Reader, out io.Writer) error {
	// TODO: use GNU readline shortcuts?
	fmt.Fprintf(out, MONKEY_FACE)
	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return fmt.Errorf("not scanned, maybe EOF")
		}
		line := scanner.Text()
		l := lexer.New(line)

		if false {
			// TODO: options
			for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
				fmt.Fprintf(out, "%+v\n", tok)
			}
		}

		p := parser.New(l)
		program := p.ParseProgram()
		if p.Error() != nil {
			printErrors(out, p.Error())
			continue
		}
		ob, err := eval.Eval(program)
		if err != nil {
			fmt.Fprintf(out, "eval err: %v", err)
			continue
		}
		// evaluate: print the well-formed AST -> flag
		// fmt.Fprintf(out, "%s\n", program.String())
		fmt.Fprintf(out, "%s\n", ob.Inspect())
	}
}

func printErrors(out io.Writer, err error) {
	io.WriteString(out, err.Error())
}
