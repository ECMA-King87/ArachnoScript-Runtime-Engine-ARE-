
# ArachnoScript Runtime Environment Change log

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