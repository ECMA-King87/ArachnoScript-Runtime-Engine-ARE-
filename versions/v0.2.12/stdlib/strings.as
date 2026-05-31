
function String(value) {
  return new String.prototype(value)
}

String.prototype = {
  string: class {
    private default str = "";
    private #length = 0;

    get length() {
      return this.#length;
    }

    constructor(value) {
      this.str = #_to_string(value);
      this.#length = #_length(this.str);
    }

    $ (index: number): string
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
      return String(#_to_uppercase(this.str))
    }

    public toLowerCase() {
      return String(#_to_lowercase(this.str))
    }
    
    $ *
    $ * replaceAll(searchValue: string | RegExp, replaceValue: string): string
    public replaceAll(searchValue, replaceValue) {
      if (typeof searchValue != "string" && !(searchValue instanceof RegExp)) throw "String.replaceAll: 1st argument must be a string or RegExp";
      if (typeof replaceValue != "string") throw "String.replaceAll: 2nd argument must be a string";
      if (searchValue instanceof RegExp) var regexp = #_value(searchValue);
      else var [regexp] = #_new_regexp(RegExp.escape(searchValue));
      return String(#_regexp_replace(this.str, regexp, replaceValue))
    }

    $ start: 0 based index of starting character
    $ end: considered as 1 based index of ending character
    public slice(start, end) {
      start ??= 0;
      end ??= -1;
      if (typeof start != "number"|| typeof end != "number") {
        throw "String.charCodeAt: arguments must be numbers";
      }

      return String(#_to_string(
        #_slice_array(
          #_new_byte_array(this.str), start, end
        )
      ))
    }

    public trim() {
      spawn new_string = this.trimStart();
      new_string = new_string.trimEnd();
      return new_string;
    }

    public trimEnd() {
      spawn new_string = String(this.str);

      while (true) {
        if (new_string.at(-1) == ('\n' | '\r' | ' ' | '\t')) new_string = new_string.slice(0, -1);
        else break;
      }

      return new_string;
    }

    public trimStart() {
      spawn new_string = String(this.str);

      while (true) {
        if (new_string.at(0) == ('\n' | '\r' | ' ' | '\t')) new_string = new_string.slice(1);
        else break;
      }

      return new_string;
    }

    $ *
    $ * (maxLength: number, fillString?: string): string;
    public repeat(count) {
      if (typeof count != "number") throw "String.repeat: argument must be of type number.";
      if (count == 0) return "";
      spawn new_string = this.str;
      for (spawn l = 0; l <= count; l++) new_string += this.str;
      return new_string;
    }

    $ *
    $ * (maxLength: number, fillString?: string): string;
    public padStart(maxLength, fillString) {
      if (typeof maxLength != "number") throw "String.padStart: 1st argument must be of type number.";
      if (typeof fillString != "string" && fillString != undefined) throw "String.padStart: 2nd argument must be of type string.";
      if (maxLength <= this.#length) return this;
      fillString ??= String(' ').repeat(maxLength-this.#length);
      // for (spawn i = #_length(this.str); i < maxLength; i++) {}
      return String(fillString + this.str);
    }

    $ *
    $ * (maxLength: number, fillString?: string): string;
    public padEnd(maxLength, fillString) {
      if (typeof maxLength != "string") throw "String.padEnd: 1st arguments must be of type number.";
      if (typeof fillString != "string" && fillString != undefined) throw "String.padEnd: 2nd arguments must be of type string.";
      if (maxLength <= this.#length) return this;
      fillString ??= String(' ').repeat(maxLength-this.#length);
      // for (spawn i = #_length(this.str); i < maxLength; i++) {}
      return String(this.str + fillString);
    }

    public valueOf(str) {
      return this.str
    }

    private [Symbol.debug]() {
      return this.str
    }

    private [Symbol.toPrimitive]() {
      return this.str
    }
  }
}.string;

String.isString = function(value) {
  return value instanceof String.prototype;
}