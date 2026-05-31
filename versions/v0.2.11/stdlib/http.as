

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
      if (typeof options != "object" && options != undefined) throw "Request: 2nd argument must be an object";
      this.#url = url;
      if (options) if (typeof options.method != "string" && options.method != undefined) throw "Request: request method must be a string";
                  else if (typeof options.method == "string") this.#method = options.method;
    }
  }
  RequestWriter: class {
    // private writer;

    constructor(private writer) {
      if (!#_is_valid_response_writer(#_value(writer))) throw "RequestWriter: 1st argument must be of type raw [request writer]";
    }

    public write(u8array) {
      if (!(u8array instanceof Uint8Array)) throw "RequestWriter.write: 1st argument must be of type Uint8Array";
      $ TODO: fix the '#_value' macro, it does not return instance default value
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

    // public listenAndServe(address) {
    //  if (typeof address != "string") throw "Server.listenAndServe: 1st argument must be a string";
    //  #_http_listen_and_serve(address, this.serve_mux);
    // }

    $ *
    $ * listenAndServe(options: { hostname: string, port: number })
    $ * options.hostname (e.g. "localhost", "127.0.0.1", "0.0.0.0")
    $ * options.port (a number above 1023)
    public listenAndServe(options) {
      if (typeof options != "object" && options != undefined) throw "Server.listenAndServe: 1st argument must be a object";
      if (typeof options.hostname != "string" && options.hostname != undefined) throw "Server.listenAndServe: invalid hostname, hostname must be a string";
      if (typeof options.port != "number" && options.port != undefined) throw "Server.listenAndServe: invalid port, port must be a number";
      options.hostname ??= ""
      options.port ??= 0
      immortal spawn listener = #_http_listen(options.hostname+":"+options.port);
      if (typeof options.onListen == "function") options.onListen({ transport: "tcp", hostname: options.hostname });
      #_http_serve(listener, this.serve_mux);
    }
  }
};