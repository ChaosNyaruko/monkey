1+2
[]
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

fn(a,b,c,f) { f(add(a,b), c) };

let array = [1,2*3,3+4,4, fn(a, b) { return a + b;}];
let x = array[1] + array[1+2];
x
let f = array[4];
f(10,2)

len(array)

first(array)
last(array)
rest(array)
rest(rest(array))
len(rest(array))

let int = [1,2,3,4];
rest(rest(rest(rest(rest(int)))))

let empty = [1];
let one = push(empty, 2);

empty
one

let map = fn(arr, f) {
	let iter = fn(res, arr) {
		if (len(arr) == 0) {
			return res;
		}
		return iter(push(res, f(first(arr))), rest(arr));
	};

	iter([], arr)
};

let a = [1,2,3,4,5];
let double = fn(x) { 2 * x };
map(a, double)

let reduce = fn(f, arr, initial) {
	let iter = fn(prev, arr) {
		if (len(arr) == 0) {
			return prev;
		}
		return iter(f(prev, first(arr)), rest(arr));
	};

	iter(initial, arr)
};

let a = [1,2,3,4,5];
let add = fn(x, y) {  x + y };
reduce(add, a, 10)

let sum = fn(arr) {
	reduce(add, arr, 0)
};
sum([1,2,3,4,5])
sum([145,46269,4524])
[]
