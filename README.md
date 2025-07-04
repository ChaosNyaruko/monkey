Learning interpreters by project.

Refs:
- [Write an Interpreter in Go](https://interpreterbook.com/)
- [The lox language](https://craftinginterpreters.com/)

# Installation
```console
$ go install github.com/ChaosNyaruko/monkey@master
```

# Examples
## Interpret a file
```console
$ monkey -f test.mk
```

## Interactive interpreter
```console
$ monkey 
```

# Syntax
See [test.mk](./test.mk).

---
The following are the main concepts involved in the [book](https://interpreterbook.com/) when I was learning to build my own monkey.

# Parser
- AST (eBNF --> (yacc, bison, antlr) -> AST)
- Write by hand, top-down.
- Expressions and Statements.

## LET
let \<Identifier> \<Expression> ;
```
let x = 5 + 5 * 10;
let y = 10;
let foobar = add(5,5);
```

## RETURN 
return \<Expression> ;
```
return 1 + 1 ;
return 10;
return x;
```

## Expressions 
### Pratt Parser
[Top Down Operator Precedence](https://tdop.github.io/)

[Pratt Parsers: Expression Parsing Made Easy ](https://journal.stuffwithstuff.com/2011/03/19/pratt-parsers-expression-parsing-made-easy/)

[Simple but Powerful Pratt Parsing](https://matklad.github.io/2020/04/13/simple-but-powerful-pratt-parsing.html)

1. precedence
A + B * C -> (A + B) * C or A + (B * C)?
```
parse(lhs, precedence, remain) {
precedence = "+" i.e. 3
lhs = "B"
for remain not end {
  op = peekprecedence() --> "*"
  if op < precedence {
    break
  } else {
    rhs = parse(op, remain[1:]) -> C
    lhs = (op lhs rhs) -> lhs = (* B C)
  }
}
lhs
}

(outside)parser() {
precedence = "" // 0 
lhs = "A"
for remain not end {
  op = nexttoken() --> "+" / 3
  if op < precedence {
    break
  } else {
    rhs = parse(op, remain[1:]) -> (* B C)
    lhs = (op lhs rhs) -> lhs (+ A (* B C))
  }
}
return lhs
}


// A*B + C

(outside)parser() {
precedence = "" // 0 
lhs = "A"

for remain not end {
  op = nexttoken() --> "+" / 3
  if op < precedence {
    break
  } else {
    rhs = inside_parse("+", "C") -> C
    lhs = (op lhs rhs) -> lhs = (* A B) -> (+ lhs C) -> (+ (* A B) C)
  }
}
return lhs
}


(inside)parser() {
precedence = "+" // 5
lhs = ""
remain = "C"

for remain not end {
  op = nexttoken() --> "+" / 3
  if op < precedence {
    break
  } else {
    rhs = parse(op, remain[1:])
    lhs = (op lhs rhs) -> lhs 
  }
}
return lhs
}
```

2. associativity
```
A + B + C   -> (A + B) + C  or A + (B + C)?

f.g.h -> f.(g.h)
(outside)parser() {
precedence = "." // 8
lhs = "f"

for remain not end {
  op = nexttoken() --> "." / (8.5, 8)
  if 8.5 < 8 {
    break
  } else {
    rhs = inside_parse("+", "C") -> C
    lhs = (op lhs rhs) -> lhs = (* A B) -> (+ lhs C) -> (+ (* A B) C)
  }
}
return lhs
}
```

3. infix/prefix/postfix

### Identifier
### Number Literal
### Prefix
\<PrefixOp> \<Expression> ;
### Infix
\<Expression> \<InfixOp> \<Expression> ;
### Boolean
"true"
"false"

### Grouping
( <Expression> )

### If-Else ExpressionStatement
if (<condition>) {<block>} else {<block>}
```
let foo = if (x>y) {x} return {y};
```
### Function Literal
fn <Params list> <Body>
<Params list> (Identifier1, Identifier2, ...)
```
let f = fn(a, b) { return a + b;}
```

### Function Calling
<Expression>(<Expression list>)
```
f(1,2)
```

# Evaluate
- tree-traversal
- byte-code: JVM
- native binary: Go/C++/C/Rust/...
## object
解释器内部对类型的某种存储
struct {
  []bit
}
“对象系统”
### Integer
### Boolean
### Null

## Eval

# More Data Structures and Builtins
## String
- let x = "hello", not support escaping
- let y = "hello" + " " + "world"
- let y = x + " " + "world"
- let l = len("hello")
## Array
```
- let myArray = [1, 2, 3, "hello", true, fn(a, b) {return a + b;}, [1,2,3]];
- myArray[0]
- myArray[6][0]
- myArray[5]()
- len(myArray)
- first(myArray) 
- last(myArray)
- rest(myArray) - a = (1 2 3) scheme: car(a) => 1 cdr(a) => (2 3)
- let myArray = push(myArray, 10);
- let myArray = map/reduce(myArray, f);
```

We can iterate now! Though we don't have a "loop" syntax.

## HashMap
- let john = {"name": "john", "age": 20, "favourite": "Marvel MCU"};
- let age = john["age"]; -> 20
- let name = john["name"]; -> "john"
- let invalid = john["invalid"]; -> null
- let invalid = john["invalid" + " other"]; -> null
- let key = "name";
- let dict = {key: "john"};
- dict[key]
- Key(object type): String/Boolean/Integer

{<Expression>: <Expression> [, <Expression>:<Expression>].*}

# Macro
similar to Exlixir

## quote / unquote 
- quote(foo) -> QUOTE(foo)
- quote(1+2) -> QUOTE((1+2))
- quote(foo + 1 + bar) -> QUOTE(((foo + 1) + bar))
- quote(foo + unquote(1 + 4)) -> QUOTE((foo + 5))

## define and use marco
- let reverse_sub = macro(a, b) {quote(unquote(b) - unquote(a))}  
  - reverse_sub(1+2, 3+4) --> (7-3) --> 4
- let first_valid_m = macro(a, b) {a}  
- let first_valid_f = fn   (a, b) {a}  
  - first_valid_m(print("first"), print("second") --> first
  - first_valid_f(print("first"), print("second") --> first second

# References
- https://www.plai.org/3/2/PLAI%20Version%203.2.2%20printing.pdf
- https://www.khoury.northeastern.edu/home/wand/eopl/
- https://book.douban.com/subject/30348061/
- https://go.dev/doc/pgo
- https://www.modular.com/mojo
- bison GLR LALR
- creeps(the game)
- PTC
- RAII (resource allocation is initialization)
- IIFE(immediately invoked function expression)

> 函数式语言就是自递归加数组模式匹配


# TODO (possibly learn them in lox)
- Easier REPL with GNU readline
- Float/Hexadecimal number support
- Error handling: no panic, report line 
- Virtual Machine
