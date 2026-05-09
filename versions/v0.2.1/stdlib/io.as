

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
  $ if the value passed is not a string or is not convertable
  $ to a byte array, then this macro will throw an error
  spawn byte_array = #_new_byte_array(string);
  #_file_write(#_os_stdout, byte_array)
}

$ *
$ * (message?: string, defaultValue?: string)?: string
function prompt(message, defaultValue) {
  immortal spawn Err = "prompt: arguments must be strings";
  if (typeof message != "string" && message != undefined) throw Err;
  if (typeof defaultValue != "string" && defaultValue != undefined) throw Err;
  if (message) {
    print(message)
  }
  $ set buffer to 1024 character size
  spawn byte_array = #_new_byte_array(1024);
  $ read from stdin and write to 'byte_array'
  #_file_read(#_os_stdin, byte_array);
  if (#_length(byte_array) == 0) return null;
  $ slice the buffer to its original length and
  $ convert the sliced buffer to a string
  return #_to_string(#_slice_array(byte_array, 0, -1));
}