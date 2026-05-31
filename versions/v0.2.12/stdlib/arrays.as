
$ *
$ * new<T>(...elements: []T): Array<T>
$ * new<T>(length: number): Array<T>
class Array {
  private #length = 0

  $ *
  $ * new<T>(next: () => { done: bool; value: T }): ArrayIterator<T>
  private ArrayIterator = class {
    constructor(next) {
      if (typeof next != "function") throw "ArrayIterator: argument must be a function";
      this[Symbol.iterator] = () => ({ next });
    }

    $ *
    $ * (): Array<T>
    toArray() {
      immortal spawn array = new Array();
      for (spawn [_, value] of this) array.push(value);
      return array;
    }

    private [Symbol.iterator]() {}
  }

  constructor(...elements) {
    spawn arrLen = #_length(elements);
    if (arrLen == 1 && typeof elements[0] == "number") {
      spawn len = elements[0];
      for (spawn i = 0; i < len; i++) {
        this[i] = null;
      }
      this.#length = len;
    } else {
      for (immortal spawn i in elements) this[i] = elements[i];
      this.#length = arrLen;
    }
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
    for (spawn i = 0; i < #_length(elements); i++) {
      this[this.#length] = elements[i];
      this.#length++;
    }
    return this.#length
  }
  $ *
  $ * Returns a new Array instance with elements from the range `start` to `end` (inclusive)
  $ * set to `value`
  $ * `start`: Index to start filling the array at. If `start` is negative, it is treated as (length+start)
  $ * `end`: Index to stop filling the array at. If `end` is negative, it is treated as (length+end)
  $ * (value: T, start: number, end: number): Array
  fill(value, start, end) {
    start ??= 0;
    end ??= -1;

    spawn startType = typeof start;
    spawn endType = typeof end;
    if (startType != "number" || endType != "number") {
      throw "Array.fill: start or end parameter is not a number; ("
        + startType + ", " + endType + ")";
    }
    end < 0 ? end += this.#length : null;
    immortal spawn array = structuredClone(this);
    for (spawn i = start; i <= end; i++) array[i] = value;
    return array
  }

  concat(array) {
    if (typeof array != "array" || !Array.isArray(array)) {
      throw "Array.concat: argument must be an array or instance of Array";
    }
    immortal spawn result = new Array();
    for (spawn i = 0; i < this.#length; i++) result.push(this[i]);
    for (spawn i = 0; i < #_length(array); i++) result.push(array[i]);
    return result;
  }

  $ *
  $ * (callback: (value: T, index: number, this: Array) => void): void
  forEach(callback) {
    if (typeof callback != "function") throw "Array.forEach: argument must be a function";
    for (spawn i = 0; i < this.#length; i++) callback(this[i], i, this);
  }

  $ *
  $ * <U>(callback: (value: T, index: number, this: Array) => U): U[]
  map(callback) {
    if (typeof callback != "function") throw "Array.forEach: argument must be a function";
    immortal spawn results = new Array();
    for (spawn i = 0; i < this.#length; i++) results.push(callback(this[i], i, this));
    return results;
  }

  $ *
  $ * (): ArrayIterator<[number, T]>
  entries() {
    spawn i = 0;
    spawn self = this;
    return new this.ArrayIterator(() => {
      return {
        done: i >= self.#length,
        value: [i, self[i++]]
      }
    })
  }

  private [Symbol.iterator]() {
    spawn i = 0;
    spawn self = this;
    return {
      next: () => {
        return {
          done: i >= self.#length,
          value: self[i++]
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
    for (spawn i = 0; i < this.#length; (i++, col++)) {
      spawn lastEl = i == this.#length - 1;
      string += #_inspect(this[i], 1) + (lastEl ? "" : ", ");
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
  return value instanceof Array
}

$ *
$ * new(value: number | array | string | raw): Uint8Array
class Uint8Array {
  $ *
  $ * Raw(byte[])
  private default buffer;
  private #length = 0;

  get length() {
    return this.#length
  }

  constructor(value) {
    switch (typeof value) {
      case "number", "array", "string": {
        this.buffer = #_new_byte_array(value)
      }
      case "raw": {
        if (#_is_byte_array(value)) {
          this.buffer = value
        } else {
          fallthrough
        }
      }
      default: throw "Uint8Array: invalid argument";
    }
    this.#length = #_length(this.buffer)
  }

  public at(index) {
    if (typeof index != "number") {
      throw "Uint8Array.at: argument must be a number";
    }
    index < 0 ? index += this.#length : 0;
    return Number(#_byte_at(this.buffer, index))
  }

  public _set(index, byte) {
    if (typeof index != "number") {
      throw "Uint8Array._set: 1st argument must be a number";
    }

    if (typeof byte == "number") byte = #_byte(byte);

    if (!#_is_byte(byte)) {
      throw "Uint8Array._set: 2nd argument must be a byte";
    }
    #_set_byte_at(this.buffer, index, byte)
  }

  [Symbol.debug](sep) {
    immortal spawn len = #_length(this.buffer);
    spawn str = "Uint8Array ("+ len +") [";
    if (len > 10) str += "\r\n";
    for (spawn i = 0; i < len; i++) {
      str += (len > 10 ? sep : "") + #_inspect(Number(#_byte_at(this.buffer, i)), 1) + (i+1 < len ? ", " : "") + (len > 10 ? "\r\n" : "");
    }
    if (len > 10) str += "\r\n";
    str += "]";
    return str;
  }
}

Object.defineProperty(Uint8Array, "_from", {
  value: $ (array: Array<number>): Uint8Array $ function(array) {
    spawn uint8array = new Uint8Array(array.length);
    for (spawn idx in array) uint8array._set(idx, array[idx]);
    return uint8array;
  }
  writable: false
})