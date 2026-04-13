function Symbol(des) {
  if (typeof des != "string") {
    throw "Symbol: argument must be a string";
  }
  return #_symbol_for(des)
}

Object.defineProperty(Symbol, "iterator", {
  value: #_symbol_for(#_iterator_symbol_descriptor),
  writable: false,
})

Object.defineProperty(Symbol, "debug", {
  value: #_symbol_for("debug"),
  writable: false,
})