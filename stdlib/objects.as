#_set_context(globalThis)

class Object {}

Object.isValidObject = function (value) {
  if (typeof value == ("function" | "object" | "instance" | "class")) {
    return true
  }
  return false
}

Object.defineProperty = function (o, p, d) {
  if (Object.isValidObject(o) && typeof d == "object") {
      #_define_property(o, p, d)
  } else {
    throw "Object.defineProperty: 1st and 3rd arguments must be objects";
  }
}

function structuredClone(value) {
  return value
}