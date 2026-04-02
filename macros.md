# Macros Reference

## Comprehensive reference for built-in macros in the ArachnoScript runtime (v0.2).

These are the macros built-in to the ArachnoScript runtime as of the 3rd of April, 2026 (v0.2)

## Overview

- Each macro is documented with: **Signature**, **Arguments**, **Returns**, and a short **Description**.
- Macros are listed alphabetically for quick lookup.

---

## Table of contents

- [#_byte](#byte)
- [#_byte_array_to_string](#byte-array-to-string)
- [#_byte_at](#byte-at)
- [#_byte_to_string](#byte-to-string)
- [#_c_free](#c-free)
- [#_c_malloc](#c-malloc)
- [#_c_strdup](#c-strdup)
- [#_define_property](#define-property)
- [#_ffi_call_function](#ffi-call-function)
- [#_ffi_get_symbol](#ffi-get-symbol)
- [#_file_close](#file-close)
- [#_file_read](#file-read)
- [#_file_size](#file-size)
- [#_file_stats](#file-stats)
- [#_file_write](#file-write)
- [#_inspect](#inspect)
- [#_is_byte](#is-byte)
- [#_is_byte_array](#is-byte-array)
- [#_length](#length)
- [#_load_shared_lib](#load-shared-lib)
- [#_main_module_path](#main-module-path)
- [#_meta_path](#meta-path)
- [#_new_byte_array](#new-byte-array)
- [#_new_error](#new-error)
- [#_new_promise](#new-promise)
- [#_open_file](#open-file)
- [#_os_write_file](#os-write-file)
- [#_parse_float](#parse-float)
- [#_parse_int](#parse-int)
- [#_parse_number](#parse-number)
- [#_path_exists](#path-exists)
- [#_path_relative](#path-relative)
- [#_push_byte](#push-byte)
- [#_queue_microtask](#queue-microtask)
- [#_slice_array](#slice-array)
- [#_symbol_for](#symbol-for)
- [#_to_number](#to-number)
- [#_to_string](#to-string)
- [#_value_is_infinity](#value-is-infinity)
- [#_value_is_nan](#value-is-nan)
- [#_write_byte_array](#write-byte-array)
---

<a id="byte"></a>
### `#_byte`
- Signature: `#_byte(value)`
- Arguments: `number | string` (single character)
- Returns: `raw [byte]`
- Description: Convert a numeric value or single-character string to a raw byte.

<a id="byte-array-to-string"></a>
### `#_byte_array_to_string`
- Signature: `#_byte_array_to_string(bytes)`
- Arguments: `raw [byte array]`
- Returns: `string`
- Description: Convert a byte array into a string.

<a id="byte-at"></a>
### `#_byte_at`
- Signature: `#_byte_at(bytes, index)`
- Arguments: `raw [byte array], number`
- Returns: `raw [byte]`
- Description: Return the byte at `index`. Negative indices count from the end.

<a id="byte-to-string"></a>
### `#_byte_to_string`
- Signature: `#_byte_to_string(byte)`
- Arguments: `raw [byte]`
- Returns: `string`
- Description: Convert a single raw byte to a one-character string.

<a id="c-free"></a>
### `#_c_free`
- Signature: `#_c_free(ptr)`
- Arguments: `raw [unsafe pointer]`
- Returns: `undefined`
- Description: Free memory allocated on the C heap using libc `free`.

<a id="c-malloc"></a>
### `#_c_malloc`
- Signature: `#_c_malloc(size)`
- Arguments: `number`
- Returns: `raw [unsafe pointer] | null`
- Description: Allocate `size` bytes on the C heap using libc `malloc`.

<a id="c-strdup"></a>
### `#_c_strdup`
- Signature: `#_c_strdup(str)`
- Arguments: `string`
- Returns: `raw [unsafe pointer] | null`
- Description: Duplicate a string onto the C heap. Uses platform `strdup` when safe; otherwise falls back to `malloc`+copy.

<a id="define-property"></a>
### `#_define_property`
- Signature: `#_define_property(object, key, descriptor)`
- Arguments: `object, key, descriptor` (object)
- Returns: `undefined`
- Description: Define a property with descriptor fields (`value`, `get`, `set`, `writeable`, `configurable`).

<a id="ffi-call-function"></a>
### `#_ffi_call_function`
- Signature: `#_ffi_call_function(fnPtr, [retType], [argTypes], ...args)`
- Arguments: `raw [fn ptr], optional return type, optional arg type array, ...args`
- Returns: depends on `retType`
- Description: Generic FFI invocation helper with optional explicit return/argument type control; otherwise attempts auto-inference.

<a id="ffi-get-symbol"></a>
### `#_ffi_get_symbol`
- Signature: `#_ffi_get_symbol(libHandle, name)`
- Arguments: `raw [shared library], string`
- Returns: `raw [function pointer]`
- Description: Look up a symbol in a loaded shared library and return a raw function pointer.

<a id="file-close"></a>
### `#_file_close`
- Signature: `#_file_close(file)`
- Arguments: `raw [os file]`
- Returns: `undefined`
- Description: Close an open file handle.

<a id="file-read"></a>
### `#_file_read`
- Signature: `#_file_read(file, buffer)`
- Arguments: `raw [os file], raw [byte array]`
- Returns: `number` (bytes read)
- Description: Read into the provided byte buffer from an open file.

<a id="file-size"></a>
### `#_file_size`
- Signature: `#_file_size(fileInfo)`
- Arguments: `raw [os file info]`
- Returns: `number`
- Description: Return the size in bytes from an `os.FileInfo` value.

<a id="file-stats"></a>
### `#_file_stats`
- Signature: `#_file_stats(path)`
- Arguments: `string`
- Returns: `raw [os file info]`
- Description: Get stat information for the resolved path.

<a id="file-write"></a>
### `#_file_write`
- Signature: `#_file_write(file, bytes)`
- Arguments: `raw [os file], raw [byte array]`
- Returns: `number` (bytes written)
- Description: Write bytes to an open file handle.

<a id="inspect"></a>
### `#_inspect`
- Signature: `#_inspect(value)`
- Arguments: `any`
- Returns: `string`
- Description: Produce a human-readable inspection string for a runtime value.

<a id="is-byte"></a>
### `#_is_byte`
- Signature: `#_is_byte(value)`
- Arguments: `any`
- Returns: `boolean`
- Description: True when the value is a raw `byte`.

<a id="is-byte-array"></a>
### `#_is_byte_array`
- Signature: `#_is_byte_array(value)`
- Arguments: `any`
- Returns: `boolean`
- Description: True when the value is a raw byte array.

<a id="length"></a>
### `#_length`
- Signature: `#_length(value)`
- Arguments: `any`
- Returns: `number`
- Description: Return a sensible length for supported runtime values (strings, arrays, objects, numbers, byte arrays); returns `-1` for unsupported types.

<a id="load-shared-lib"></a>
### `#_load_shared_lib`
- Signature: `#_load_shared_lib(path)`
- Arguments: `string`
- Returns: `raw [shared library handle]`
- Description: Load a shared library and return a handle usable by other FFI macros.

<a id="main-module-path"></a>
### `#_main_module_path`
- Signature: `#_main_module_path()`
- Arguments: none
- Returns: `string`
- Description: Return the interpreter's main source module path.

<a id="meta-path"></a>
### `#_meta_path`
- Signature: `#_meta_path()`
- Arguments: none
- Returns: `string`
- Description: Return the current scope/module path.

<a id="new-byte-array"></a>
### `#_new_byte_array`
- Signature: `#_new_byte_array([spec])`
- Arguments: none | `number` | `string` | `array of bytes/numbers`
- Returns: `raw [byte array]`
- Description: Construct a new byte array from the provided specification.

<a id="new-error"></a>
### `#_new_error`
- Signature: `#_new_error(name, message)`
- Arguments: `string name, string message`
- Returns: `raw [Error]`
- Description: Create a runtime Error object including a stack trace.

<a id="new-promise"></a>
### `#_new_promise`
- Signature: `#_new_promise(callback)`
- Arguments: `function`
- Returns: `promise`
- Description: Create a new promise from a callback function.

<a id="open-file"></a>
### `#_open_file`
- Signature: `#_open_file(path)`
- Arguments: `string`
- Returns: `raw [os file]`
- Description: Open a file by absolute path and return a handle.

<a id="os-write-file"></a>
### `#_os_write_file`
- Signature: `#_os_write_file(path, bytes, mode)`
- Arguments: `string, raw [byte array], number`
- Returns: `null | throws`
- Description: Write bytes to a filesystem path with given file mode.

<a id="parse-float"></a>
### `#_parse_float`
- Signature: `#_parse_float(str)`
- Arguments: `string`
- Returns: `number | NaN`
- Description: Parse a string as a floating-point number.

<a id="parse-int"></a>
### `#_parse_int`
- Signature: `#_parse_int(str)`
- Arguments: `string`
- Returns: `number | NaN`
- Description: Parse a string as an integer (base auto-detected).

<a id="parse-number"></a>
### `#_parse_number`
- Signature: `#_parse_number(str)`
- Arguments: `string`
- Returns: `number | NaN`
- Description: Parse a string to a float or int; tries float first then int.

<a id="path-exists"></a>
### `#_path_exists`
- Signature: `#_path_exists(path)`
- Arguments: `string`
- Returns: `boolean`
- Description: Check whether a filesystem path exists.

<a id="path-relative"></a>
### `#_path_relative`
- Signature: `#_path_relative(path)`
- Arguments: `string`
- Returns: `string`
- Description: Resolve a path relative to the current scope and return the real path.

<a id="push-byte"></a>
### `#_push_byte`
- Signature: `#_push_byte(bytes, byte)`
- Arguments: `raw [byte array], raw [byte]`
- Returns: `raw [byte array]`
- Description: Append a byte to a byte array and return the array.

<a id="queue-microtask"></a>
### `#_queue_microtask`
- Signature: `#_queue_microtask(fn)`
- Arguments: `function`
- Returns: `undefined`
- Description: Schedule a microtask to call the provided function.

<a id="slice-array"></a>
### `#_slice_array`
- Signature: `#_slice_array(arrayOrBytes, start, end)`
- Arguments: `array | raw [byte array], number start, number end`
- Returns: `array | raw [byte array]`
- Description: Return a slice supporting negative indices.

<a id="symbol-for"></a>
### `#_symbol_for`
- Signature: `#_symbol_for(name)`
- Arguments: `string`
- Returns: `symbol`
- Description: Intern and return a symbol for `name`.

<a id="to-number"></a>
### `#_to_number`
- Signature: `#_to_number(value)`
- Arguments: `any`
- Returns: `number | NaN`
- Description: Convert a runtime value to a number when possible.

<a id="to-string"></a>
### `#_to_string`
- Signature: `#_to_string(value)`
- Arguments: `any`
- Returns: `string`
- Description: Convert a runtime value to its string representation.

<a id="value-is-infinity"></a>
### `#_value_is_infinity`
- Signature: `#_value_is_infinity(value)`
- Arguments: `number`
- Returns: `boolean`
- Description: True when the numeric value is infinite.

<a id="value-is-nan"></a>
### `#_value_is_nan`
- Signature: `#_value_is_nan(value)`
- Arguments: `number`
- Returns: `boolean`
- Description: True when the numeric value is NaN.

<a id="write-byte-array"></a>
### `#_write_byte_array`
- Signature: `#_write_byte_array(src, dst, pos)`
- Arguments: `raw [byte array], raw [byte array], number`
- Returns: `raw [byte array]`
- Description: Copy bytes from `src` into `dst` at `pos` and return `dst`.

---

If you want, I can now:

- Add short usage examples for selected macros.
- Generate a linked table-of-contents (with working anchors) for HTML rendering.


