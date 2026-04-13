
class Promise {
  private default promise;
  constructor(callback) {
    if (typeof callback != "function") {
      throw "TypeError: Promise: 1st argument must be a function"
      // throw #_new_error("TypeError", "Promise: 1st argument must be a function");
    }
    this.promise = #_new_promise(callback);
  }

  [#_symbol_for("debug")]() {
    return #_inspect(this.promise)
  }

  then(callback) {
    if (typeof callback != "function") {
      throw "TypeError: Promise.then: 1st argument must be a function"
      // throw #_new_error("TypeError", "Promise.then: 1st argument must be a function");
    }
    this.promise.then(callback)
  }

  catch(callback) {
    if (typeof callback != "function") {
      throw "TypeError: Promise.catch: 1st argument must be a function"
    }
    this.promise.catch(callback)
  }
}

