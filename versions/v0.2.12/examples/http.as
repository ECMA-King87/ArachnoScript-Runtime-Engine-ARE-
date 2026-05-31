
var server = new Http.Server();

server.handle("/", (w, r) => {
  Console.log(r.method, r.url);
  w.write(new Uint8Array("Hello World!"))
});

@coroutine
server.listenAndServe({
  hostname: "localhost",
  $ automatically choos the port number
  port: 0,
  $ the port could not be exposed for the tcp listener
  onListen(localAddr) {
    Console.log("Hello World!, listening on", localAddr.transport, "http://"+localAddr.hostname)
  }
});

