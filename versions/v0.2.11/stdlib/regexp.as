


class RegExp {
  private #regexp;
  constructor(exp) {
    if (typeof exp != "string") throw "RegExp: argument must be a string";
    immortal spawn [regexp, err] = #_new_regexp(exp);
    if (err) throw "RegExp: " + err;
    this.#regexp = regexp;
  }

  public test(text) {
    return #_regexp_test(this.#regexp, text)
  }
}

RegExp.escape = (text) => {
  if (typeof text != "string") throw "RegExp.escape: argument must be a string";
  return #_regexp_escape(text)
};