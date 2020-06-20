# Hot Reloading

This is a **experimental** prototype of Hot Reload for controllerbus.

Features:

 - Load + resolve controllerbus controllers from plugins
 - Manifest of all currently loaded plugins
 - Detect Go module code changes and hot-reload

Cavets:

 - Go plugin system cannot fully unload memory, making a mem leak.
 - Plugins only work on Linux.

These are acceptable for development environments. Memory leak issue could be
addressed with ControllerBus IPC using cross-process controllers.

