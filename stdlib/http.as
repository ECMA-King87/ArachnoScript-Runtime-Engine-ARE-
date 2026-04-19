#_set_context(globalThis)

static spawn Http = {
  Request: class {
    private #method = "GET";
    private #url = "";

    public get method() {
      return this.#method;
    }

    public get url() {
      return this.#url;
    }

    constructor(url, options) {
      if (typeof url != "string") throw "Request: 1st argument must be a string";
      if (typeof options != "object" && options != undefined) throw "Request: 1st argument must be an object";
      this.#url = url;
      $ TODO: add nullish coalasce feature
      // options.method ?? this.#method = options.method;
      if (typeof options.method != "string") throw "Request: request method must be a string";
      options.method ? this.#method = options.method : 0;
    }
  }
  RequestWriter: class {
    // private writer;

    constructor(private writer) {
      if (#_is_valid_response_writer(writer)) throw "RequestWriter: 1st argument must be of type raw [request writer]";
    }

    public write(u8array) {
      if (!(u8array instanceof Uint8Array)) throw "RequestWriter.write: 1st argument must be of type Uint8Array";
      return this.writer.write(#_value(u8array));
    }
  }
  Server: class {
    private serve_mux = #_new_http_serve_mux();

    constructor() {}

    public handle(path, handler) {
      if (typeof path != "string") throw "Server.handle: 1st argument must be a string";
      if (typeof handler != "function") throw "Server.handle: 2nd argument must be a function";

      #_serve_mux_handle_func(this.serve_mux, path, (w, r) => {
        handler(new Http.RequestWriter(w), new Http.Request(r.url, { method: r.method }));
      });
    }

    public listenAndServe(port) {
      if (typeof port != "number") throw "Server.listenAndServe: 1st argument must be a number";
      #_http_listen_and_serve(":" + port, this.serve_mux);
    }
  }
};

var server = new Http.Server();

server.handle("/", (w, r) => {
  Console.log("new request!")
});

immortal spawn port = 4567;

@coroutine
server.listenAndServe(port);

Console.log("Hello World!, listening on http://localhost:"+port)