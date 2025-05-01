let map = fn(arr, f) {
	let iter = fn(res, arr) {
		if (len(arr) == 0) {
			return res;
		}
		return iter(push(res, f(first(arr))), rest(arr));
	};

	iter([], arr)
};
