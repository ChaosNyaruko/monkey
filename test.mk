1+2
let a = 1000;
let b = 2000;
a + b;

let c = if (a > b) { "a is BIGGER than b: hello" } else { "a is SMALLER than b: world"};
c
let s = "hello world";
len(s)
let s = "hello world\n";
len(s)

let add = fn(a, b) { return a + b;};
let mul = fn(a,b) { return a * b;};
let ternary = fn(a,b,c,f) { f(add(a,b), c) };
ternary(1,2,10,mul);

