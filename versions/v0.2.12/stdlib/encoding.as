static spawn TextEncoding = {
  encode(str) {
    if (typeof str != "string") throw "TextEncoding.encode: argument must be a string.";
    return new Uint8Array(str);
  }
  decode(buffer) {
    if (!(buffer instanceof Uint8Array)) throw "TextEncoding.decode: argument must be a Uint8Array.";
    return String(#_value(buffer))
  }
};