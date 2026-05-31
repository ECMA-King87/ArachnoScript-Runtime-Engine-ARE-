# Macros Reference

Comprehensive reference for built-in macros in the ArachnoScript runtime (v0.2.x).

These are the macros built into the ArachnoScript runtime as declared in `src/macros.go`.

## Overview

- Each macro entry includes: Signature, Arguments, Returns and a concise Description.
- FFI-related macros are experimental; review their descriptions before use.

---

## Table of contents

- [#_value_is_nan](#value-is-nan)
- [#_os_args](#os-args)
- [#_to_string](#to-string)
- [#_value](#value)
- [#_to_number](#to-number)
- [#_to_uppercase](#to-uppercase)
- [#_new_regexp](#new-regexp)
- [#_regexp_escape](#regexp-escape)
- [#_regexp_test](#regexp-test)
- [#_regexp_replace](#regexp-replace)
- [#_to_lowercase](#to-lowercase)
- [#_value_is_infinity](#value-is-infinity)
- [#_parse_number](#parse-number)
- [#_parse_float](#parse-float)
- [#_parse_int](#parse-int)
- [#_symbol_for](#symbol-for)
- [#_symbol_keyfor](#symbol-keyfor)
- [#_byte](#byte)
- [#_new_promise](#new-promise)
- [#_queue_microtask](#queue-microtask)
- [#_os_write_file](#os-write-file)
- [#_load_shared_lib](#load-shared-lib)
- [#_free_shared_lib](#free-shared-lib)
- [#_ffi_get_symbol](#ffi-get-symbol)
- [#_c_malloc](#c-malloc)
- [#_c_free](#c-free)
- [#_c_strdup](#c-strdup)
- [#_ffi_call_function](#ffi-call-function)
- [#_file_write](#file-write)
- [#_file_read](#file-read)
- [#_new_byte_array](#new-byte-array)
- [#_write_byte_array](#write-byte-array)
- [#_push_byte](#push-byte)
- [#_byte_at](#byte-at)
- [#_set_byte_at](#set-byte-at)
- [#_is_byte_array](#is-byte-array)
- [#_is_byte](#is-byte)
- [#_slice_array](#slice-array)
- [#_open_file](#open-file)
- [#_path_exists](#path-exists)
- [#_file_stats](#file-stats)
- [#_file_size](#file-size)
- [#_path_relative](#path-relative)
- [#_path_relative_to_file](#path-relative-to-file)
- [#_real_path](#real-path)
- [#_file_close](#file-close)
- [#_http_handle](#http-handle)
- [#_new_http_serve_mux](#new-http-serve-mux)
- [#_http_handle_func](#http-handle-func)
- [#_serve_mux_handle_func](#serve-mux-handle-func)
- [#_serve_mux_handle](#serve-mux-handle)
- [#_http_listen](#http-listen)
- [#_http_serve](#http-serve)
- [#_is_valid_response_writer](#is-valid-response-writer)
- [#_length](#length)
- [#_new_error](#new-error)
- [#_define_property](#define-property)
- [#_meta_path](#meta-path)
- [#_main_module_path](#main-module-path)
- [#_inspect](#inspect)
- [#_worker](#worker)
- [#_set_context](#set-context)
- [#_sqrt](#sqrt)
- [#_sine](#sine)
- [#_cosine](#cosine)
- [#_tangent](#tangent)
- [#_arcsine](#arcsine)
- [#_arccosine](#arccosine)
- [#_arctangent](#arctangent)
- [#_log](#log)
- [#_absolute](#absolute)
- [#_arctangent2](#arctangent2)
- [#_ceil](#ceil)
- [#_floor](#floor)
- [#_round](#round)
- [#_random](#random)
- [#_max](#max)
- [#_min](#min)
- [#_runtime_version](#runtime_version)
- [#_sleep](#sleep)
- [#_time](#time)
- [#_get_millisec](#get-millisec)
- [#_get_second](#get-second)
- [#_get_minute](#get-minute)
- [#_get_hour](#get-hour)
- [#_get_date](#get-date)
- [#_get_weekday](#get-weekday)
- [#_get_month](#get-month)
- [#_get_year](#get-year)
- [#_get_time_loc](#get-time-loc)
- [#_unix_milli](#unix-milli)
- [#_assert](#assert)

---

<a id="value-is-nan"></a>
### `#_value_is_nan`
- Signature: `#_value_is_nan(value)`
- Arguments: `number`
- Returns: `boolean`
- Description: Returns true when the provided numeric value is NaN.

<a id="os-args"></a>
### `#_os_args`
- Signature: `#_os_args()`
- Arguments: none
- Returns: `array [string]`
- Description: Returns the process command-line arguments as an array of strings (`os.Args`).

<a id="to-string"></a>
### `#_to_string`
- Signature: `#_to_string(value)`
- Arguments: `any`
- Returns: `string`
- Description: Convert a runtime value to its string representation. Raw byte arrays and raw bytes are converted to strings by interpreting their bytes.

<a id="value"></a>
### `#_value`
- Signature: `#_value(value)`
- Arguments: `any`
- Returns: `any`
- Description: If given an object-like value, searches its own properties and prototype chain for a Default property and returns that value; otherwise returns the argument unchanged.

<a id="to-number"></a>
### `#_to_number`
- Signature: `#_to_number(value)`
- Arguments: `any`
- Returns: `number | NaN`
- Description: Convert a value to a number. Supports `NumberVal` and raw `byte` values; returns `NaN` for unsupported types.

<a id="to-uppercase"></a>
### `#_to_uppercase`
- Signature: `#_to_uppercase(str)`
- Arguments: `string`
- Returns: `string`
- Description: Return the uppercase form of the provided string (falls back to the value's string representation if not a string).

<a id="new-regexp"></a>
### `#_new_regexp`
- Signature: `#_new_regexp(pattern)`
- Arguments: `string`
- Returns: `array [raw [regexp], null|string]` (regexp handle and optional error string)
- Description: Compile a regular expression pattern; returns `[regexp, null]` on success or `[regexp, errorMessage]` on failure.

<a id="regexp-escape"></a>
### `#_regexp_escape`
- Signature: `#_regexp_escape(str)`
- Arguments: `string`
- Returns: `string`
- Description: Escape regexp metacharacters in `str` (wrapper around `regexp.QuoteMeta`).

<a id="regexp-test"></a>
### `#_regexp_test`
- Signature: `#_regexp_test(re, str)`
- Arguments: `raw [regexp], string`
- Returns: `boolean`
- Description: Test whether `str` matches the compiled `regexp`.

<a id="regexp-replace"></a>
### `#_regexp_replace`
- Signature: `#_regexp_replace(str, re, replacement)`
- Arguments: `string, raw [regexp], string`
- Returns: `string`
- Description: Replace matches of `re` in `str` with `replacement`.

<a id="to-lowercase"></a>
### `#_to_lowercase`
- Signature: `#_to_lowercase(str)`
- Arguments: `string`
- Returns: `string`
- Description: Return the lowercase form of the provided string (falls back to the value's string representation if not a string).

<a id="value-is-infinity"></a>
### `#_value_is_infinity`
- Signature: `#_value_is_infinity(value)`
- Arguments: `number`
- Returns: `boolean`
- Description: Returns true when the numeric value is infinite.

<a id="parse-number"></a>
### `#_parse_number`
- Signature: `#_parse_number(str)`
- Arguments: `string`
- Returns: `number | NaN`
- Description: Attempts to parse a string as a float first, falling back to integer parse; returns `NaN` if neither parse succeeds.

<a id="parse-float"></a>
### `#_parse_float`
- Signature: `#_parse_float(str)`
- Arguments: `string`
- Returns: `number | NaN`
- Description: Parse a string as a floating-point number; returns `NaN` on failure.

<a id="parse-int"></a>
### `#_parse_int`
- Signature: `#_parse_int(str)`
- Arguments: `string`
- Returns: `number | NaN`
- Description: Parse a string as an integer (auto-detects base); returns `NaN` on failure.

<a id="symbol-for"></a>
### `#_symbol_for`
- Signature: `#_symbol_for(name)`
- Arguments: `string`
- Returns: `symbol`
- Description: Return an interned symbol for the given name (uses the runtime symbol table).

<a id="symbol-keyfor"></a>
### `#_symbol_keyfor`
- Signature: `#_symbol_keyfor(symbol)`
- Arguments: `symbol`
- Returns: `string | undefined`
- Description: Return the string key for a registered symbol, or `undefined` if the symbol is not found in the symbol table.

<a id="byte"></a>
### `#_byte`
- Signature: `#_byte(value)`
- Arguments: `number | string` (single character)
- Returns: `raw [byte]`
- Description: Convert a numeric value or single-character string to a raw byte.

<a id="new-promise"></a>
### `#_new_promise`
- Signature: `#_new_promise(callback)`
- Arguments: `function`
- Returns: `promise`
- Description: Create a new promise from the provided callback function.

<a id="queue-microtask"></a>
### `#_queue_microtask`
- Signature: `#_queue_microtask(fn)`
- Arguments: `function`
- Returns: `undefined`
- Description: Schedule the provided function to run as a microtask.

<a id="os-write-file"></a>
### `#_os_write_file`
- Signature: `#_os_write_file(path, bytes, mode)`
- Arguments: `string, raw [byte array], number` (file mode)
- Returns: `null | throws`
- Description: Write bytes to the filesystem at `path` (resolved relative to the provided base); `mode` is interpreted as an unsigned int file mode; errors raise a FileWriteError.

<a id="load-shared-lib"></a>
### `#_load_shared_lib`
- Signature: `#_load_shared_lib(path)`
- Arguments: `string`
- Returns: `raw [shared library handle]`
- Description: Load a shared library (using FFI) and return a handle usable by other FFI macros.

<a id="free-shared-lib"></a>
### `#_free_shared_lib`
- Signature: `#_free_shared_lib(libHandle)`
- Arguments: `raw [shared library]`
- Returns: `undefined | throws`
- Description: Free a previously loaded shared library handle; errors raise a Warning source error.

<a id="ffi-get-symbol"></a>
### `#_ffi_get_symbol`
- Signature: `#_ffi_get_symbol(libHandle, name)`
- Arguments: `raw [shared library], string`
- Returns: `raw [function pointer]`
- Description: Look up a symbol name in a loaded shared library and return a raw function pointer.

<a id="c-malloc"></a>
### `#_c_malloc`
- Signature: `#_c_malloc(size)`
- Arguments: `number`
- Returns: `raw [unsafe pointer] | null`
- Description: Allocate `size` bytes on the C heap via libc `malloc`. Returns `null` if allocation fails.

<a id="c-free"></a>
### `#_c_free`
- Signature: `#_c_free(ptr)`
- Arguments: `raw [unsafe pointer]`
- Returns: `undefined | throws`
- Description: Free C memory allocated on the libc heap using `free`.

<a id="c-strdup"></a>
### `#_c_strdup`
- Signature: `#_c_strdup(str)`
- Arguments: `string`
- Returns: `raw [unsafe pointer] | null`
- Description: Duplicate a Go string onto the C heap. Uses platform `strdup` when available and safe; otherwise falls back to `malloc`+copy. Returns `null` on allocation failure.

<a id="ffi-call-function"></a>
### `#_ffi_call_function`
- Signature: `#_ffi_call_function(fnPtr, [retType], [argTypes], ...args)`
- Arguments: `raw [unsafe pointer], optional return type (string), optional arg type array (array of strings), ...args`
- Returns: `depends on retType`
- Description: Generic FFI invocation helper. When `retType` and/or `argTypes` are omitted the macro tries to infer types from the provided arguments. Supported explicit types include pointers, strings, booleans, integers, floats, and more; see runtime usage in `src/macros.go` for exact mappings.

<a id="file-write"></a>
### `#_file_write`
- Signature: `#_file_write(file, bytes)`
- Arguments: `raw [os file], raw [byte array]`
- Returns: `number` (bytes written)
- Description: Write the provided bytes to an open `os.File` handle; raises FileWriteError on failure.

<a id="file-read"></a>
### `#_file_read`
- Signature: `#_file_read(file, buffer)`
- Arguments: `raw [os file], raw [byte array]`
- Returns: `number` (bytes read)
- Description: Read from an open `os.File` into the provided byte buffer; raises FileReadError on failure.

<a id="new-byte-array"></a>
### `#_new_byte_array`
- Signature: `#_new_byte_array([spec])`
- Arguments: none | `number` | `string` | `array` | varargs of `raw [byte]`/`number`
- Returns: `raw [byte array]`
- Description: Construct a new byte array. Accepts multiple forms: no args -> empty array; a single number -> allocate that length; string -> byte contents; an array or list of bytes/numbers -> builds byte array from elements.

<a id="write-byte-array"></a>
### `#_write_byte_array`
- Signature: `#_write_byte_array(src, dst, pos)`
- Arguments: `raw [byte array] src, raw [byte array] dst, number pos`
- Returns: `raw [byte array]` (dst)
- Description: Copy the contents of `src` into `dst` beginning at index `pos`, and return `dst`.

<a id="push-byte"></a>
### `#_push_byte`
- Signature: `#_push_byte(bytes, byte)`
- Arguments: `raw [byte array], raw [byte]`
- Returns: `raw [byte array]`
- Description: Append a single raw byte to a byte array and return the array.

<a id="byte-at"></a>
### `#_byte_at`
- Signature: `#_byte_at(bytes, index)`
- Arguments: `raw [byte array], number`
- Returns: `raw [byte]`
- Description: Return the byte at `index`. Negative indices count from the end; errors on out-of-range access.

<a id="set-byte-at"></a>
### `#_set_byte_at`
- Signature: `#_set_byte_at(bytes, index, byte)`
- Arguments: `raw [byte array], number, raw [byte]`
- Returns: `raw [byte array]`
- Description: Set the byte at `index` to the provided byte value. Negative indices count from the end; errors on out-of-range access. Returns the modified byte array.

<a id="is-byte-array"></a>
### `#_is_byte_array`
- Signature: `#_is_byte_array(value)`
- Arguments: `any`
- Returns: `boolean`
- Description: True when the value is a raw byte array.

<a id="is-byte"></a>
### `#_is_byte`
- Signature: `#_is_byte(value)`
- Arguments: `any`
- Returns: `boolean`
- Description: True when the value is a raw byte.

<a id="slice-array"></a>
### `#_slice_array`
- Signature: `#_slice_array(arrayOrBytes, start, end)`
- Arguments: `array | raw [byte array], number start, number end`
- Returns: `array | raw [byte array]`
- Description: Return a slice of an array or byte array. Supports negative indices which count from the end.

<a id="open-file"></a>
### `#_open_file`
- Signature: `#_open_file(path)`
- Arguments: `string`
- Returns: `raw [os file]`
- Description: Open a file at the given path and return an `os.File` handle.

<a id="path-exists"></a>
### `#_path_exists`
- Signature: `#_path_exists(path)`
- Arguments: `string`
- Returns: `boolean`
- Description: Check whether a filesystem path exists.

<a id="file-stats"></a>
### `#_file_stats`
- Signature: `#_file_stats(path)`
- Arguments: `string`
- Returns: `raw [os file info]`
- Description: Return `os.FileInfo` for the resolved path; raises PathError on failure.

<a id="file-size"></a>
### `#_file_size`
- Signature: `#_file_size(fileInfo)`
- Arguments: `raw [os file info]`
- Returns: `number`
- Description: Return the file size in bytes from an `os.FileInfo` value.

<a id="path-relative"></a>
### `#_path_relative`
- Signature: `#_path_relative(base, target)`
- Arguments: `string base, string target`
- Returns: `string`
- Description: Compute a relative path from `base` to `target`.

<a id="path-relative-to-file"></a>
### `#_path_relative_to_file`
- Signature: `#_path_relative_to_file(file, target)`
- Arguments: `string file, string target`
- Returns: `string`
- Description: Resolve `target` relative to `file` and return the relative path.

<a id="real-path"></a>
### `#_real_path`
- Signature: `#_real_path(path)`
- Arguments: `string`
- Returns: `string`
- Description: Return the real (resolved/absolute) filesystem path for `path`.

<a id="file-close"></a>
### `#_file_close`
- Signature: `#_file_close(file)`
- Arguments: `raw [os file]`
- Returns: `undefined`
- Description: Close an open file handle.

---

### HTTP macros

<a id="http-handle"></a>
### `#_http_handle`
- Signature: `#_http_handle(pattern, handler)`
- Arguments: `string pattern, raw [http.Handler] handler`
- Returns: `undefined`
- Description: Register a raw `http.Handler` for `pattern` on the default server mux.
- Example:
```
#_http_handle("/static", myHandlerRaw)
```

<a id="new-http-serve-mux"></a>
### `#_new_http_serve_mux`
- Signature: `#_new_http_serve_mux()`
- Arguments: none
- Returns: `raw [*http.ServeMux]`
- Description: Create and return a new `ServeMux` instance (raw pointer wrapped by the runtime).
- Example:
```
let mux = #_new_http_serve_mux()
#_serve_mux_handle_func(mux, "/ok", (w, req) => { w.write(#_new_byte_array("OK")) })
```

<a id="http-handle-func"></a>
### `#_http_handle_func`
- Signature: `#_http_handle_func(pattern, fn)`
- Arguments: `string pattern, function fn(writer, request)`
- Returns: `undefined`
- Description: Register a runtime function as an `http.HandlerFunc` on the default server mux. The handler receives two arguments: a `writer` object (has `write` helper) and a `request` object (with `method`, `url`).
- Example:
```
#_http_handle_func("/hello", (w, req) => {
	w.write(#_new_byte_array("Hello, world\n"))
})
```

<a id="serve-mux-handle-func"></a>
### `#_serve_mux_handle_func`
- Signature: `#_serve_mux_handle_func(mux, pattern, fn)`
- Arguments: `raw [*http.ServeMux] mux, string pattern, function fn(writer, request)`
- Returns: `undefined`
- Description: Register a runtime function handler on the provided `ServeMux` instance.
- Example:
```
let mux = #_new_http_serve_mux()
#_serve_mux_handle_func(mux, "/greet", (w, req) => { w.write(#_new_byte_array("Hi")) })
```

<a id="serve-mux-handle"></a>
### `#_serve_mux_handle`
- Signature: `#_serve_mux_handle(mux, pattern, handler)`
- Arguments: `raw [*http.ServeMux] mux, string pattern, raw [http.Handler] handler`
- Returns: `undefined`
- Description: Register a raw `http.Handler` on a `ServeMux` instance.
- Example:
```
#_serve_mux_handle(mux, "/static", myHandlerRaw)
```

<a id="http-listen"></a>
### `#_http_listen`
- Signature: `#_http_listen(address)`
- Arguments: `string address` (e.g. `"127.0.0.1:8080"`)
- Returns: `raw [net.Listener]`
- Description: Bind to the given TCP address and return a listener. Throws an `HttpError` on bind failure.
- Example:
```
let ln = #_http_listen("127.0.0.1:8080")
```

<a id="http-serve"></a>
### `#_http_serve`
- Signature: `#_http_serve(listener, handler)`
- Arguments: `raw [net.Listener] listener, raw [http.Handler] | raw [*http.ServeMux] | undefined handler`
- Returns: `undefined` (this call blocks while serving)
- Description: Start serving HTTP requests on `listener` using `handler`. If `handler` is `undefined`, the default `http.DefaultServeMux` is used.
- Example (serve a mux):
```
#_http_serve(ln, mux)
```

<a id="is-valid-response-writer"></a>
### `#_is_valid_response_writer`
- Signature: `#_is_valid_response_writer(value)`
- Arguments: `any`
- Returns: `boolean`
- Description: Return `true` when `value` is a runtime-wrapped `http.ResponseWriter`.
- Example:
```
#_is_valid_response_writer(someVal) // -> true|false
```

---

<a id="length"></a>
### `#_length`
- Signature: `#_length(value)`
- Arguments: `any`
- Returns: `number`
- Description: Return a sensible length for supported runtime values (numbers, strings, arrays, objects, byte arrays); returns `0` for unsupported types.

<a id="new-error"></a>
### `#_new_error`
- Signature: `#_new_error(name, message)`
- Arguments: `string name, string message`
- Returns: `raw [Error]`
- Description: Construct a runtime `Error` object with the provided name and message and include the current stack trace.

<a id="define-property"></a>
### `#_define_property`
- Signature: `#_define_property(object, key, descriptor)`
- Arguments: `object, key, descriptor (object)`
- Returns: `undefined`
- Description: Define a property on `object` using the provided descriptor object. Descriptor fields: `value`, `get`, `set`, `writable`, `configurable`, `public`.

<a id="meta-path"></a>
### `#_meta_path`
- Signature: `#_meta_path()`
- Arguments: none
- Returns: `string`
- Description: Return the current scope/module path.

<a id="main-module-path"></a>
### `#_main_module_path`
- Signature: `#_main_module_path()`
- Arguments: none
- Returns: `string`
- Description: Return the interpreter's main source module path.

<a id="inspect"></a>
### `#_inspect`
- Signature: `#_inspect([value])`
- Arguments: `any` (optional)
- Returns: `string`
- Description: Return a human-readable inspection string for the provided runtime value. If omitted, returns the inspection of `undefined`.

<a id="worker"></a>
### `#_worker`
- Signature: `#_worker(modulePath, runInParallel)`
- Arguments: `string modulePath, boolean runInParallel`
- Returns: `undefined`
- Description: Load and execute the specified module. If `runInParallel` is true the module runs on the global work queue; otherwise it runs on the current thread.

<a id="set-context"></a>
### `#_set_context`
- Signature: `#_set_context(scopeObject)`
- Arguments: `object [scope object]`
- Returns: `undefined`
- Description: Replace the current execution scope with the provided scope object.

<a id="get-context"></a>
### `#_get_context`
- Signature: `#_get_context()`
- Returns: `object [scope object]`
- Description: Returns the current execution scope as a scope object.

---

### Math macros

<a id="sqrt"></a>
### `#_sqrt`
- Signature: `#_sqrt(number)`
- Returns: `number`
- Description: Square root.

<a id="sine"></a>
### `#_sine`
- Signature: `#_sine(number)`
- Returns: `number`
- Description: Sine.

<a id="cosine"></a>
### `#_cosine`
- Signature: `#_cosine(number)`
- Returns: `number`
- Description: Cosine.

<a id="tangent"></a>
### `#_tangent`
- Signature: `#_tangent(number)`
- Returns: `number`
- Description: Tangent.

<a id="arcsine"></a>
### `#_arcsine`
- Signature: `#_arcsine(number)`
- Returns: `number`
- Description: Arcsine.

<a id="arccosine"></a>
### `#_arccosine`
- Signature: `#_arccosine(number)`
- Returns: `number`
- Description: Arccosine.

<a id="arctangent"></a>
### `#_arctangent`
- Signature: `#_arctangent(number)`
- Returns: `number`
- Description: Arctangent.

<a id="log"></a>
### `#_log`
- Signature: `#_log(number)`
- Returns: `number`
- Description: Natural logarithm.

<a id="absolute"></a>
### `#_absolute`
- Signature: `#_absolute(number)`
- Returns: `number`
- Description: Absolute value.

<a id="arctangent2"></a>
### `#_arctangent2`
- Signature: `#_arctangent2(y, x)`
- Returns: `number`
- Description: Two-argument arctangent (atan2).

<a id="ceil"></a>
### `#_ceil`
- Signature: `#_ceil(number)`
- Returns: `number`
- Description: Ceiling.

<a id="floor"></a>
### `#_floor`
- Signature: `#_floor(number)`
- Returns: `number`
- Description: Floor.

<a id="round"></a>
### `#_round`
- Signature: `#_round(number)`
- Returns: `number`
- Description: Round to nearest integer.

<a id="random"></a>
### `#_random`
- Signature: `#_random()`
- Returns: `number`
- Description: Return a pseudorandom float between 0 and 1.

<a id="max"></a>
### `#_max`
- Signature: `#_max(a, ...b)`
- Returns: `number`
- Description: Return the maximum of the provided numeric arguments.

<a id="min"></a>
### `#_min`
- Signature: `#_min(a, ...b)`
- Returns: `number`
- Description: Return the minimum of the provided numeric arguments.

<a id="runtime_version"></a>
### `#_runtime_version`
- Signature: `#_runtime_version()`
- Returns: `string`
- Description: Return the version number of the running ArachnoScript Runtime.

<a id="sleep"></a>
### `#_sleep`
- Signature: `#_sleep(milliseconds)`
- Arguments: `number`
- Returns: `undefined`
- Description: Pause execution for the specified number of milliseconds.

<a id="time"></a>
### `#_time`
- Signature: `#_time([year, month, day, hours, minutes, seconds, milliseconds])` or `#_time(dateString)`
- Arguments: optional - `number, number, number, number, number, number, number` or `string` (RFC3339 format)
- Returns: `raw [time object]`
- Description: Return the current time as a raw time object. If arguments are provided, construct a time object from the given date/time components (month is 1-indexed). If a date string is provided, parse it in RFC3339 format.

<a id="get-millisec"></a>
### `#_get_millisec`
- Signature: `#_get_millisec(timeObject)`
- Arguments: `raw [time object]`
- Returns: `number`
- Description: Get the millisecond component (0-999) from a time object.

<a id="get-second"></a>
### `#_get_second`
- Signature: `#_get_second(timeObject)`
- Arguments: `raw [time object]`
- Returns: `number`
- Description: Get the second component (0-59) from a time object.

<a id="get-minute"></a>
### `#_get_minute`
- Signature: `#_get_minute(timeObject)`
- Arguments: `raw [time object]`
- Returns: `number`
- Description: Get the minute component (0-59) from a time object.

<a id="get-hour"></a>
### `#_get_hour`
- Signature: `#_get_hour(timeObject)`
- Arguments: `raw [time object]`
- Returns: `number`
- Description: Get the hour component (0-23) from a time object.

<a id="get-date"></a>
### `#_get_date`
- Signature: `#_get_date(timeObject)`
- Arguments: `raw [time object]`
- Returns: `number`
- Description: Get the day of month (1-31) from a time object.

<a id="get-weekday"></a>
### `#_get_weekday`
- Signature: `#_get_weekday(timeObject)`
- Arguments: `raw [time object]`
- Returns: `string`
- Description: Get the day of week name (e.g., "Monday", "Tuesday") from a time object.

<a id="get-month"></a>
### `#_get_month`
- Signature: `#_get_month(timeObject)`
- Arguments: `raw [time object]`
- Returns: `number`
- Description: Get the month (1-12) from a time object.

<a id="get-year"></a>
### `#_get_year`
- Signature: `#_get_year(timeObject)`
- Arguments: `raw [time object]`
- Returns: `number`
- Description: Get the year from a time object.

<a id="get-time-loc"></a>
### `#_get_time_loc`
- Signature: `#_get_time_loc(timeObject)`
- Arguments: `raw [time object]`
- Returns: `string`
- Description: Get the timezone location name from a time object.

<a id="unix-milli"></a>
### `#_unix_milli`
- Signature: `#_unix_milli(timeObject)`
- Arguments: `raw [time object]`
- Returns: `number`
- Description: Get the Unix time in milliseconds since epoch from a time object.

<a id="assert"></a>
### `#_assert`
- Signature: `#_assert(condition)`
- Arguments: `boolean`
- Returns: `undefined | throws`
- Description: Assert that the given condition is true. If the condition is false, throws an `Assertion` error with the current stack trace.



