import "objects.as"
import "symbols.as"
import "math.as"
import "io.as"
import "fs.as"
import "promise.as"
import "arrays.as"
import "numbers.as"
import "strings.as"

var arr = new Array();

var mux = #_new_http_serve_mux();

#_serve_mux_handle_func(mux, "/", (w, r) => {
  Console.log("new request!")
});

@coroutine
#_http_listen_and_serve(":4567", mux)

Console.log("Hello World!, listening on http://localhost:4567")

@debug String("Super").slice(0, 3)


@benchmark for (i = 0; i < 10; i++);
@coroutine {
  for (i = 0; i < 10; i++);
  Console.log("done for");
}

spawn i = 0;
@benchmark while ((i < 10, i++));
@coroutine {
  while ((i < 10, i++));
  Console.log("done while");
}