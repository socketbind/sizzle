# Sizzle

Shitty little hack which plays sizzling sounds according to CPU usage. Spawns the specified command as a subprocess and watches the CPU usage  including its children.

Usage:
```
sizzle <cmd> <cmd args>
```

Example:
```
sizzle make -j4
```

## Installation
```
go get github.com/socketbind/sizzle

$GOPATH/bin/sizzle <args> # or put the executable whereever you wish
```
