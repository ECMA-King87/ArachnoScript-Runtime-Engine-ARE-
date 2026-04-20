#_set_context(globalThis)

function String(value) {
  return new (class {
    private default str = "";
    constructor(value) {
      this.str = #_to_string(value);
    }

    public at(index) {
      if (typeof index != "number") {
        throw "String.at: index must be a number"
      }
      if (index < 0) {
        index += this.#length
      }
      return this.str[index]
    }

    public charCodeAt(index) {
      if (typeof index != "number") {
        throw "String.charCodeAt: argument must be a number";
      }
      return Number(#_byte(this.str[index]))
    }

    $ start: 0 based index of starting character
    $ end: considered as 1 based index of ending character
    public slice(start, end) {
      start ??= 0;
      end ??= -1;
      if (typeof start != "number"|| typeof end != "number") {
        throw "String.charCodeAt: arguments must be numbers";
      }

      return #_to_string(
        #_slice_array(
          #_new_byte_array(this.str), start, end
        )
      )
    }

    private [Symbol.debug]() {
      return this.str
    }
  })(value)
}
