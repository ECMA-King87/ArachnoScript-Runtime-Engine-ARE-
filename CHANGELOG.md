
# ArachnoScript Runtime Change log

## ArachnoScript v0.2.12 Change log

### Parser
- updated for loop parsing.
- updated object literal parsing.
- updated arrow function syntax.
- proposed that the 'default' keyword in property delclarations will be removed in later versions.

### Runtime
- updated 'arrays' standard library module.
- updated `#_inspect` macro.
- updated `#_worker` macro.
- updated `#_symbol_for` macro.
- fixed 'io' standard library module bug.
- fixed array data type debug printing.
- fixed bug with resolving 'this' object in function calls.
- updated 'arrays', and 'strings' standard library modules.
- fixed runtime error bug.
- added `#_set_context` macro.
- added `#_sleep` macro.
- added `#_time` macro.
- added `#_unix_milli` macro.
- added `#_get_year` macro.
- added `#_get_month` macro.
- added `#_get_date` macro.
- added `#_get_weekday` macro.
- added `#_get_hour` macro.
- added `#_get_minute` macro.
- added `#_get_second` macro.
- added `#_get_millisec` macro.
- added `#_symbol_keyfor` macro.
- added `#_assert` macro.
- added 'encoding', and 'date' standard library modules.
- handled division by zero operation.

## ArachnoScript v0.2.11 Change log

### Runtime
- fixed for loop bug
- fixed file association in installer for running scripts
- marked `#_ffi_call_function` macro as unstable.
- fixed multithread bug with micro tasks.
- updated 'runtime' standard library module.
- updated for loop syntax

## ArachnoScript v0.2.1 Change log

### Lexer
- optimized lexer.
- fixed lexer bugs in string lexing.

### Parser
- updated constructor logic to match runtime behavior.
- updated object literal parsing.
- updated if statement parsing.
- fixed many parser bugs.
- optimized class property parsing logic
- updated increment/decrement nodes

### Runtime
- enabled variable declaration in if statement condition.
- fixed bug in `#_new_byte_array` macro.
- updated `#_serve_mux_handle_func` macro.
- updated `#_new_byte_array` macro.
- fixed `#_path_relative` macro.
- removed `#_http_listen_and_serve` macro.
- added `#_http_listen` macro
- added `#_http_serve` macro
- added `#_path_relative` macro
- added `#_real_path` macro
- added `#_path_relative_to_file` macro
- updated 'strings', 'arrays', 'promise', 'io', 'fs', 'math', and 'main' standard library modules.
- added 'http', and 'regexp' standard library modules.
- added documentations to 'arrays' standard library module.
- updated documentations to 'io' standard library module.
- updated scope objects feature.
- updated object data type property descriptor.
- updated class property initializer logic
- updated scope logic
- updated map lookups
- updated import statement logic
- updated for in/of loop
- updated object data type logic
- fixed await logic/bug
- updated `#_worker` macro.
- updated macro data type error architecture.


## ArachnoScript v0.2.0 Change log

### Lexer
- updated lexer logic (moved from using regular expressions).

### Parser
- updated parser logic (using precedence helpers).
- optimized parser (nodes, logic)

### Runtime
- improved interpreter logic
- improved errors (Parse errors and runtime errors).
- improved array & object destructuring features.
- improved import statement feature.
- improved 'static' variable declaration functions.
- improved object data type logic.
- improved function call and return keyword logic.
- optimized interpreter heap.
- optimized 'while' and 'for' loops.
- fixed bugs in certain data types.
- fixed bugs in traditional for loops.
- updated native builtin class logic and function async/wait.
- updated event loop.
- updated 'io', 'fs', 'runtime', 'string', 'symbols', 'numbers', 'arrays', and 'objects' standard library modules.
- updated control flow logic (throw, continue, break).
- updated 'var' keyword functions.
- updated class declaration logic.
- updated 'this' variable functions.
- updated 'globalThis' keywords parsing and functions.
- updated 'default'  modifier for class properties.
- updated switch case logic.
- updated member assignment logic.
- updated 'main' entry point program.
- removed timed microtasks.
- removed builtin REPL (Read Eval Print Loop).
- removed original label feature.
- removed '#_print' macro.
- removed 'byte arrays', 'characters' standard library modules.
- added scope object data type.
- added property descriptors to object data type.
- added the new label language feature.
- added 'fallthrough' keyword.
- added 'module' variable at module scope.
- added getter & setter modifiers in classes.
- added 'RangeError', stack overflow errors.
- added '#_inspect' macro, and many others.
- enabled string escapes.
- enabled line and doc comment parsing.
- enabled multiple declarations separated by commas syntax.
- enabled private and public modifiers' full functionality.
- enabled tuple equality check.
- enabled ffi (Foreign Function Interface) through dedicated macros (Experimental).
- changed class constructor method syntax (removed function keyword).
