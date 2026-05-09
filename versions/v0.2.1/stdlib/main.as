$ sets the current scope/environment to it's argument which
$ is a scope object, and in this case, the scope object
$ is the globalThis object
#_set_context(globalThis)

import "objects.as"
import "symbols.as"
import "math.as"
import "io.as"
import "fs.as"
import "promise.as"
import "arrays.as"
import "numbers.as"
import "regexp.as"
import "strings.as"
import "http.as"
import "runtime.as"

immortal spawn argsLength = #_length(Runtime.args);

if (argsLength > 1) {

  for (spawn i = 0; i < argsLength; i++) {
    spawn arg = Runtime.args[i];
    spawn nextArg = "";
    if (i+1 < argsLength) {
      nextArg = Runtime.args[i+1]
    }
    switch (arg) {
      case "-e": (#_worker(nextArg, false), i++)
      case "run": {
        if (nextArg == "") {
          throw "An argument to 'run' must be given to execute script."
        }
        $ creates a new worker
        $ the 2nd argument specifies whether to run in a separate thread (true) or block the current thread (false)
        #_worker(nextArg, false)
      }
    }
  }
} else {
  Console.log("ArachnoScript Runtime Environment (ARE) \x1b[32mv0.2.1\x1b[0m");
  Console.log("          _____                    _____          ");
  Console.log("         /\\    \\                  /\\    \\         ");
  Console.log("        /::\\    \\                /::\\    \\        ");
  Console.log("       /::::\\    \\              /::::\\    \\       ");
  Console.log("      /::::::\\    \\            /::::::\\    \\      ");
  Console.log("     /:::/\\:::\\    \\          /:::/\\:::\\    \\     ");
  Console.log("    /:::/__\\:::\\    \\        /:::/__\\:::\\    \\    ");
  Console.log("   /::::\\   \\:::\\    \\       \\:::\\   \\:::\\    \\   ");
  Console.log("  /::::::\\   \\:::\\    \\    ___\\:::\\   \\:::\\    \\  ");
  Console.log(" /:::/\\:::\\   \\:::\\    \\  /\\   \\:::\\   \\:::\\    \\ ");
  Console.log("/:::/  \\:::\\   \\:::\\____\\/::\\   \\:::\\   \\:::\\____\\");
  Console.log("\\::/    \\:::\\  /:::/    /\\:::\\   \\:::\\   \\::/    /");
  Console.log(" \\/____/ \\:::\\/:::/    /  \\:::\\   \\:::\\   \\/____/ ");
  Console.log("          \\::::::/    /    \\:::\\   \\:::\\    \\     ");
  Console.log("           \\::::/    /      \\:::\\   \\:::\\____\\    ");
  Console.log("           /:::/    /        \\:::\\  /:::/    /    ");
  Console.log("          /:::/    /          \\:::\\/:::/    /     ");
  Console.log("         /:::/    /            \\::::::/    /      ");
  Console.log("        /:::/    /              \\::::/    /       ");
  Console.log("        \\::/    /                \\::/    /        ");
  Console.log("         \\/____/                  \\/____/         ");
  Console.log("                                                  ");

  input()
}