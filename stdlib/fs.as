
$ returns *RawVal[[]byte] (raw [byte array])
function write_file(path, data) {
  // if (!(data instanceof Uint8Array)) {}
  if (!#_is_byte_array(data)) {
    throw #_new_error("TypeError", "write_file: 2nd argument is not a byte array")
  }
  var file = #_open_file(path);
  var b = #_file_write(file, data);
  #_file_close(file);
  return b
}

$ returns *RawVal[[]byte] (raw [byte array])
function read_file(path) {
  $path = #_path_relative(path)
  var file = #_open_file(path);
  var stats = #_file_stats(path);
  var size = #_file_size(stats);
  $ todo: wrap with Uint8Array
  var buffer = #_new_byte_array(size);
  #_file_read(file, buffer);
  #_file_close(file)
  return buffer
}

function read_text_file(path) {
  return #_byte_array_to_string(read_file(path));
}