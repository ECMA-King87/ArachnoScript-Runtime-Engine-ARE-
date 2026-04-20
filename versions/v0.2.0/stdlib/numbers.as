#_set_context(globalThis)

function Number(value) {
  if (typeof value == "string") {
    return #_parse_number(value)
  } else if (typeof value == "number") {
    return value
  } else if (#_is_byte(value)) {
    return #_to_number(value)
  }
  return NaN
}

Number.isInteger = function (value) {
  if (typeof value != "number") return false;
  if (value % 1 == 0) return true;
  return false;
}

Number.isNaN = function (value) {
  // NaN == NaN; // false
  // NaN != NaN; // true
  return value != value
}

Number.isSafeInteger = function (value) {
  if (typeof value != "number") {
    return false
  }
  return value < Number.MAX_SAFE_INTEGER
}

Number.parseFloat = function (string) {
  if (typeof string != "string") {
    throw "Number.parseFloat: expects it's argument to be of type string"
  }
  return #_parse_float(string)
}

Number.parseInt = function (string) {
  if (typeof string != "string") {
    throw "Number.parseInt: expects it's argument to be of type string"
  }
  return #_parse_int(string)
}

Number.isFinite = function (number) {
  if (typeof number != "number") {
    throw "Number.parseInt: expects it's argument to be of type number"
  }
  return number < Infinity
}

Object.defineProperty(Number, "MAX_VALUE", {
  value: #_max_number_value, // 1.79E+308
  writeable: false,
})

Object.defineProperty(Number, "MAX_SAFE_INTEGER", {
  value: #_max_safe_integer, // 9007199254740991 2^53 − 1
  writeable: false,
})

Object.defineProperty(Number, "MIN_SAFE_INTEGER", {
  value: #_min_safe_integer, // 9007199254740991 2^53 − 1
  writeable: false,
})

Object.defineProperty(Number, "MIN_VALUE", {
  value: #_min_number_value, // 
  writeable: false,
})

Object.defineProperty(Number, "NaN", {
  value: NaN, // 
  writeable: false,
})

Object.defineProperty(Number, "POSITIVE_INFINITY", {
  value: Infinity, // a number greater than the largest number representable
  writeable: false,
})


// TODO: Number.EPSILON; // 2.2204460492503130808472633361816 x 10‍−‍16