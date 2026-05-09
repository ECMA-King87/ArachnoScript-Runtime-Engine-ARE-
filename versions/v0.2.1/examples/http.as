
var server = new Http.Server();

server.handle("/", (w, r) => {
  Console.log("new request!", w, r);
  $ Http.RequestWriter.write is unstable!
  // w.write(new Uint8Array("Hello World!"))
});

immortal spawn port = 4567;

@coroutine
server.listenAndServe(port);

Console.log("Hello World!, listening on http://localhost:"+port)