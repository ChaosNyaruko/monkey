let map = fn(arr, f) {
	let iter = fn(res, arr) {
		if (len(arr) == 0) {
			return res;
		}
		return iter(push(res, f(first(arr))), rest(arr));
	};

	iter([], arr)
};

let y = map([1,2,3,4,5], fn(x) { 2 * x });
print(y);

let reduce = fn(f, arr, initial) {
	let iter = fn(prev, arr) {
		if (len(arr) == 0) {
			return prev;
		}
		return iter(f(prev, first(arr)), rest(arr));
	};

	iter(initial, arr)
};
let sum = reduce(fn(a,b) { a + b }, [1,2,3,4,5], 0);
print(sum);
