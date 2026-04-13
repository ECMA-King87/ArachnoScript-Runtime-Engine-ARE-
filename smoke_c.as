// Smoke test for C FFI macros
// Allocates, strdup's, and frees memory via libc
var p = #_c_malloc(16);
var s = #_c_strdup("hello world");
// free both pointers
#_c_free(p);
#_c_free(s);
Console.log("smoke_c: ok");
