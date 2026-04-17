# Macros Reference

Comprehensive reference for built-in macros in the ArachnoScript runtime (v0.2).

These are the macros built in-to to the ArachnoScript runtime.

## Overview

- Each macro entry includes: Signature, Arguments, Returns and a concise Description.
- FFI-related macros are experimental; read their descriptions carefully before use.

---

## Table of contents

- [#_value_is_nan](#value-is-nan)
- [#_os_args](#os-args)
- [#_to_string](#to-string)
- [#_value](#value)
- [#_to_number](#to-number)
- [#_value_is_infinity](#value-is-infinity)
- [#_parse_number](#parse-number)
- [#_parse_float](#parse-float)
- [#_parse_int](#parse-int)
- [#_symbol_for](#symbol-for)
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
- [#_byte_to_string](#byte-to-string)
- [#_byte_array_to_string](#byte-array-to-string)
- [#_is_byte_array](#is-byte-array)
- [#_is_byte](#is-byte)
- [#_slice_array](#slice-array)
- [#_open_file](#open-file)
- [#_path_exists](#path-exists)
- [#_file_stats](#file-stats)
- [#_file_size](#file-size)
- [#_path_relative](#path-relative)
- [#_file_close](#file-close)
- [#_length](#length)
- [#_new_error](#new-error)
- [#_define_property](#define-property)
- [#_meta_path](#meta-path)
- [#_main_module_path](#main-module-path)
- [#_inspect](#inspect)
- [#_worker](#worker)
- [#_set_context](#set-context)

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
- Description: Convert a runtime value to its string representation. If the value is a raw byte array it is converted to a string by interpreting the bytes directly.

<a id="value"></a>
### `#_value`
- Signature: `#_value(value)`
- Arguments: `any`
- Returns: `any`
- Description: If given an `Instance`, attempts to resolve a default property value from its prototype and return that; otherwise returns the argument unchanged.

<a id="to-number"></a>
### `#_to_number`
- Signature: `#_to_number(value)`
- Arguments: `any`
- Returns: `number | NaN`
- Description: Convert a value to a number. Supports `NumberVal` and raw `byte` values; returns `NaN` for unsupported types.

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
- Description: Write bytes to the filesystem at `path` (resolved relative to the current scope). `mode` is interpreted as an unsigned int file mode; errors raise a FileWriteError.

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

<a id="byte-to-string"></a>
### `#_byte_to_string`
- Signature: `#_byte_to_string(byte)`
- Arguments: `raw [byte]`
- Returns: `string`
- Description: Convert a single raw byte into a one-character string.

<a id="byte-array-to-string"></a>
### `#_byte_array_to_string`
- Signature: `#_byte_array_to_string(bytes)`
- Arguments: `raw [byte array]`
- Returns: `string`
- Description: Convert a raw byte array into a string by interpreting its bytes.

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
- Description: Open a file at the given path (resolved to an absolute path) and return an `os.File` handle.

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
- Signature: `#_path_relative(path)`
- Arguments: `string`
- Returns: `string`
- Description: Resolve a path relative to the current scope and return the real (absolute) path.

<a id="file-close"></a>
### `#_file_close`
- Signature: `#_file_close(file)`
- Arguments: `raw [os file]`
- Returns: `undefined`
- Description: Close an open file handle.

<a id="length"></a>
### `#_length`
- Signature: `#_length(value)`
- Arguments: `any`
- Returns: `number`
- Description: Return a sensible length for supported runtime values (numbers, strings, arrays, objects, byte arrays); returns `-1` for unsupported types.

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
- Description: Define a property on `object` using the provided descriptor object. Descriptor fields: `value`, `get`, `set`, `writeable`, `configurable`, `public`.

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

---

If you want the canonical ordering, descriptions, or examples changed, tell me which macros you'd like examples for and I can add short usage snippets.

