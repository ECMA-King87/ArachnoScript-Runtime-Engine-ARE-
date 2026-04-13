static spawn Console = {
  log(...args) {
    var str = "";
    for (var arg of args) {
      str += #_inspect(arg) + " "
    }
    #_file_write(#_os_stdout,
        #_new_byte_array(str+"\r\n")
      )
  }
};

$ a simple print function that prints strings and only strings
function print(string) {
  // if the value passed is not a string or is not convertable
  // to a byte array, then this macro will throw an error
  spawn byte_array = #_new_byte_array(string);
  #_file_write(#_os_stdout, byte_array)
}

$ a simple input function that reads from the standard
$ output
function input() {
  $ set buffer to 1024 character size
  spawn byte_array = #_new_byte_array(1024);
  $ read from stdin and write to input
  #_file_read(#_os_stdin, byte_array);
  $ converts the buffer to a string
  return #_byte_array_to_string(#_slice_array(byte_array, 0, -1));
}