

function String(value) {
  return new (class {
    private default str = "";
    constructor(value) {
      this.str = #_to_string(value);
    }

    public at(index) {
      if (typeof index != "number") throw "String.at: index must be a number";
      if (index < 0) {
        index += this.#length
      }
      return this.str[index]
    }

    public charCodeAt(index) {
      if (typeof index != "number") throw "String.charCodeAt: argument must be a number";
      return Number(#_byte(this.str[index]))
    }

    public toUpperCase() {
      return #_to_uppercase(this.str)
    }

    public toLowerCase() {
      return #_to_lowercase(this.str)
    }
    
    $ *
    $ * replaceAll(searchValue: string | RegExp, replaceValue: string): string
    public replaceAll(searchValue, replaceValue) {
      if (typeof searchValue != "string" && !(searchValue instanceof RegExp)) throw "String.replaceAll: 1st argument must be a string or RegExp";
      if (typeof replaceValue != "string") throw "String.replaceAll: 2nd argument must be a string";
      if (searchValue instanceof RegExp) var regexp = #_value(searchValue);
      else var [regexp] = #_new_regexp(RegExp.escape(searchValue));
      return #_regexp_replace(this.str, regexp, replaceValue)
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
