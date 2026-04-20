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
import "strings.as"
import "runtime.as"

if (#_length(runtime.args) > 0) {
  var path = runtime.args[1];
  $ creates a new worker
  $ the 2nd argument is a flag that states whether to run in separate thread (true)
  $ or block current thread (false)
  #_worker(path, false)
} else {
  Console.log("ArachnoScript Runtime Environment (ARE) \x1b[32mv0.2.0\x1b[0m");
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