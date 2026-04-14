
// Example: call native strlen and allocate/free memory via C allocator
var lib = #_load_shared_lib("msvcrt.dll");

var strlen = #_ffi_get_symbol(lib, "strlen");

// Example 1: allocate a C string and pass pointer (avoids caller-side conversion)
var s1 = #_c_strdup("Super!");
var len1 = #_ffi_call_function(strlen, "number", s1);
Console.log("strlen (direct strdup):\t", len1);
#_c_free(s1);

// Example 2: allocate C string with strdup, pass pointer to strlen, then free
#_c_free(s)
var s = #_c_strdup("Hello from AS");
var len2 = #_ffi_call_function(strlen, "number", ["ptr"], s);
Console.log("strlen (strdup+ptr):\t", len2);
#_c_free(s);

#_free_shared_lib(lib)
