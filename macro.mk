let reverse_sub = macro(a, b) {quote(unquote(b) - unquote(a))};
let ms = reverse_sub(1+2, 3+4);
print(ms);
