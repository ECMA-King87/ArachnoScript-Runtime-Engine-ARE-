
static spawn Math = {
  $ Pi. This is the ratio of the circumference of a circle to its diameter.
  // PI: 22 / 7,
  $ The mathematical constant e. This is Euler's number, the base of natural logarithms.
  // E: ~2.7183
  $ The natural logarithm of 10.
  // LN10: number
  $ The natural logarithm of 2.
  // LN2: number
  $ The base 2 logarithm of e.
  // LOG2E: number,
  $ The base 10 logarithm of e.
  // LOG10E: number,
  $ The square root of 0.5, or, equivalently, one divided by the square root of 2.
  // SQRT1_2: number,
  $ The square root of 2.
  // SQRT2: 1.4142,

  // PHI: number,

  $ *
  $ * Returns the absolute value of a number (the value without regard to whether it is positive or negative).
  $ * For example, the absolute value of -5 is the same as the absolute value of 5.
  $ * @param x A numeric expression for which the absolute value is needed.
  $ * (x: number): number;
  abs(x) {
    if (typeof x != "number") throw "Math.abs: argument must be a number";
    return #_absolute(x)
  }
  $ *
  $ * Returns the arc cosine (or inverse cosine) of a number.
  $ * @param x A numeric expression.
  $ * (x: number): number;
  acos(x) {
    if (typeof x != "number") throw "Math.acos: argument must be a number";
    return #_arccosine(x)
  }
  $ *
  $ * Returns the arcsine of a number.
  $ * @param x A numeric expression.
  $ * (x: number): number;
  asin(x) {
    if (typeof x != "number") throw "Math.asin: argument must be a number";
    return #_arcsine(x)
  }
  $ *
  $ * Returns the arctangent of a number.
  $ * @param x A numeric expression for which the arctangent is needed.
  $ * (x: number): number;
  atan(x) {
    if (typeof x != "number") throw "Math.atan: argument must be a number";
    return #_arctangent(x)
  }
  $ *
  $ * Returns the angle (in radians) between the X axis and the line going through both the origin and the given point.
  $ * @param y A numeric expression representing the cartesian y-coordinate.
  $ * @param x A numeric expression representing the cartesian x-coordinate.
  $ * (y: number, x: number): number;
  atan2(y, x) {
    if (typeof y != "number") throw "Math.atan2: argument must be a number";
    if (typeof x != "number") throw "Math.atan2: argument must be a number";
    return #_arctangent2(y, x)
  }
  $ *
  $ * Returns the smallest integer greater than or equal to its numeric argument.
  $ * @param x A numeric expression.
  $ * (x: number): number;
  ceil(x) {
    if (typeof x != "number") throw "Math.ceil: argument must be a number";
    return #_ceil(x)
  }
  $ *
  $ * Returns the cosine of a number.
  $ * @param x A numeric expression that contains an angle measured in radians.
  $ * (x: number): number;
  cos(x) {
    if (typeof x != "number") throw "Math.cos: argument must be a number";
    return #_cosine(x)
  }
  $ *
  $ * Returns e (the base of natural logarithms) raised to a power.
  $ * @param x A numeric expression representing the power of e.
  $ * (x: number): number;
  exp(x) {
    if (typeof x != "number") throw "Math.exp: argument must be a number";
    return Math.E ** x
  }
  $ *
  $ * Returns the greatest integer less than or equal to its numeric argument.
  $ * @param x A numeric expression.
  $ * (x: number): number;
  floor(x) {
    if (typeof x != "number") throw "Math.floor: argument must be a number";
    return #_floor(x)
  }
  $ *
  $ * Returns the natural logarithm (base e) of a number.
  $ * @param x A numeric expression.
  $ * (x: number): number;
  log(x) {
    if (typeof x != "number") throw "Math.log: argument must be a number";
    return #_log(x)
  }
  $ *
  $ * Returns the larger of a set of supplied numeric expressions.
  $ * @param values Numeric expressions to be evaluated.
  $ * (...values: number[]): number;
  max(...values) {
    for (var i = 0; i < values.length; i++) {
      if (typeof values[i] != "number") throw "Math.max: all arguments must be numbers";
    }
    return #_max(...values)
  }
  $ *
  $ * Returns the smaller of a set of supplied numeric expressions.
  $ * @param values Numeric expressions to be evaluated.
  $ * (...values: number[]): number;
  min(...values) {
    for (var i = 0; i < values.length; i++) {
      if (typeof values[i] != "number") throw "Math.min: all arguments must be numbers";
    }
    return #_min(...values)
  }
  $ *
  $ * Returns the value of a base expression taken to a specified power.
  $ * @param x The base value of the expression.
  $ * @param y The exponent value of the expression.
  $ * (x: number, y: number): number;
  pow(x, y) {
    if (typeof x != "number") throw "Math.pow: argument must be a number";
    if (typeof y != "number") throw "Math.pow: argument must be a number";
    return x ** y
  }
  $ *
  $ * Returns a pseudorandom number between 0 and 1.
  $ * (): number;
  random() { return #_random() }
  $ *
  $ * Returns a supplied numeric expression rounded to the nearest integer.
  $ * @param x The value to be rounded to the nearest integer.
  $ * (x: number): number;
  round(x) {
    if (typeof x != "number") throw "Math.round: argument must be a number";
    return #_round(x)
  }
  $ *
  $ * Returns the sine of a number.
  $ * @param x A numeric expression that contains an angle measured in radians.
  $ * (x: number): number;
  sin(x) {
    if (typeof x != "number") throw "Math.sin: argument must be a number";
    return #_sine(x)
  }
  $ *
  $ * Returns the square root of a number.
  $ * @param x A numeric expression.
  $ * (x: number): number;
  sqrt(x) {
    if (typeof x != "number") throw "Math.sqrt: argument must be a number";
    return #_sqrt(x)
  }
  $ *
  $ * Returns the tangent of a number.
  $ * @param x A numeric expression that contains an angle measured in radians.
  $ * (x: number): number;
  tan(x) {
    if (typeof x != "number") throw "Math.tan: argument must be a number";
    return #_tangent(x)
  }
};

{
  immortal spawn PI = 3.14159265358979323846264338327950288419716939937510582097494459;
  Object.defineProperty(Math, "PI", {
    writable: false,
    value: PI,
  })

  immortal spawn PHI = 1.61803398874989484820458683436563811772030917980576286213544862;
  Object.defineProperty(Math, "PHI", {
    writable: false,
    value: PHI,
  })

  immortal spawn E = 2.71828182845904523536028747135266249775724709369995957496696763;
  Object.defineProperty(Math, "E", {
    writable: false,
    value: E,
  })

  immortal spawn SQRT2 = 1.41421356237309504880168872420969807856967187537694807317667974;
  Object.defineProperty(Math, "SQRT2", {
    writable: false,
    value: SQRT2,
  })

  immortal spawn SQRT1_2 = 1 / SQRT2;
  Object.defineProperty(Math, "SQRT1_2", {
    writable: false,
    value: SQRT1_2,
  })

  immortal spawn LN10 = 2.30258509299404568401799145468436420760110148862877297603332790;
  Object.defineProperty(Math, "LN10", {
    writable: false,
    value: LN10,
  })

  immortal spawn LN2 = 0.693147180559945309417232121458176568075500134360255254120680009;
  Object.defineProperty(Math, "LN2", {
    writable: false,
    value: LN2,
  })

  immortal spawn LOG10E = 1 / LN10;
  Object.defineProperty(Math, "LOG10E", {
    writable: false,
    value: LOG10E,
  })

  immortal spawn LOG2E = 1 / LN2;
  Object.defineProperty(Math, "LOG2E", {
    writable: false,
    value: LOG2E,
  })
}