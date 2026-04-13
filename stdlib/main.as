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
