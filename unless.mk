let unless_b = macro(cond, a, b) {
	if (!(cond)) {
		a
	} else {
		b
	}
};
unless_b(10 > 5, print("holy shxxt, 10 is not greater than 5!"), print("yes!"))

print("--func unless, will print both branches--")
let unless_f = fn(cond, a, b) {
	if (!cond) {
		a
	} else {
		b
	}
};


unless_f(10 > 5, print("holy shxxt, 10 is not greater than 5!"), print("yes!"))
print("------macro unless, will only print one branch--------")

let unless = macro(cond, a, b) {
quote(
	if (!unquote(cond)) {
		unquote(a)
	} else {
		unquote(b)
	}
)
};
unless(10 > 5, print("holy shxxt, 10 is not greater than 5!"), print("yes!"))

