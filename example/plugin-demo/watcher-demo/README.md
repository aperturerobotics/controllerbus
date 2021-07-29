# Watcher Demo

[![asciicast](https://asciinema.org/a/427731.svg)](https://asciinema.org/a/427731)

You can run this demo with "go run -trimpath -v ./"

This will analyze the packages (same as plugin-demo), compile the plugin with a
build hash in the filename, and watch the relevant code files for changes.

If you change any of the compiled code files, the compiler will automatically
re-assemble a new plugin binary with the updated packages.
