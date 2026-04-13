class Uint8Array {
  private default buffer;
  constructor(value) {
    switch (typeof value) {
      case "number", "array": {
        this.buffer = #_new_byte_array(value)
      }
      case "raw": {
        if (#_is_byte_array) {
          this.buffer = value
        } else {
          fallthrough
        }
      }
      default: throw "Uint8Array: invalid argument";
    }
  }

  public at(index) {
    if (typeof index != "number") {
      throw "Promise.at: argument must be a number";
    }
    index < 0 ? index += this.#length : 0;
    return Number(#_byte_at(this.buffer, index))
  }

  [Symbol.debug](sep) {
    immortal spawn len = #_length(this.buffer);
    var str = "Uint8Array ("+ len +") [";
    if (len > 10) str += "\r\n";
    for (i = 0; i < len; i++) {
      str += (len > 10 ? sep : "") + Number(#_byte_at(this.buffer, i)) + (i+1 < len ? ", " : "") + (len > 10 ? "\r\n" : "");
    }
    if (len > 10) str += "\r\n";
    str += "]";
    return str;
  }
}

class Array {
  private #length = 0
  constructor(...elements) {
    for (immortal spawn i in elements) {
      this[i] = elements[i]
    }
    this.#length = #_length(elements)
  }

  get length() {
    return this.#length
  }

  at(index) {
    if (typeof index != "number") {
      throw "Array.at: index must be a number"
    }
    if (index < 0) {
      index += this.#length
    }
    return this[index]
  }

  push(...elements) {
    for (i = 0; i < #_array_length(elements); i++) {
      this[this.#length + i] = elements[i];
      this.#length++;
    }
  }

  fill(value, start, end) {
    start ??= 0;
    end ??= -1;

    var startType = typeof start;
    var endType = typeof end;
    if (startType != "number" || endType != "number") {
      throw "Array.fill: start or end parameter is not a number; ("
        + startType + ", " + endType + ")";
    }
    end < 0 ? end += this.#length : null;
    immortal spawn array = structuredClone(this);
    for (i = start; i <= end; i++) {
      array[i] = value;
    }
    return array
  }

  concat(array) {
    if (typeof array != "array" || !Array.isArray(array)) {
      throw "Array.concat: argument must be an array or instance of Array";
    }
    throw "Array.concat: unimplemented";
  }

  private [Symbol.iterator]() {
    spawn i = 0;
    // spawn self = this;
    return {
      next: () => {
        return {
          done: i >= this.#length,
          value: this[i++]
        }
      }
    }
  }

  private [Symbol.debug](sep) {
    spawn col = 1;
    spawn string = "[";
    spawn greaterThan5 = this.#length > 5;
    if (greaterThan5) {
      string += "\r\n" + sep;
    }
    for (i = 0; i < this.#length; (i++, col++)) {
      spawn lastEl = i == this.#length - 1;
      string += #_inspect(this[i]) + (lastEl ? "" : ", ");
      if (greaterThan5 && col == 5) {
        string += "\r\n" + sep;
        col = 1;
      }
      if (greaterThan5 && lastEl) {
        string += "\r\n";
      }
    }
    return string + "]";
  }
}

Array.isArray = function (value) {
  if (value instanceof Array) {
    return !0
  }
  return !1
}
