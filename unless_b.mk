let unless_b = macro(cond, a, b) {
	if (!(cond)) {
		a
	} else {
		b
	}
};
unless_b(10 > 5, print("holy shxxt, 10 is not greater than 5!"), print("yes!"))
