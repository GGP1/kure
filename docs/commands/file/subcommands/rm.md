## Use

`kure file rm <names>`

## Description

Remove files or directories.

## Flags 

No flags.

#### Goroutines

Goroutines can be thought of as light weight threads managed by the Go runtime. There might be only one thread in a program with thousands of Goroutines.

The cost of creating a Goroutine is tiny when compared to a thread, while the minimum stack size is defined as 2048 bytes, the Go runtime does also not allow goroutines to exceed a maximum stack size; this maximum depends on the architecture and is 1 GB for 64-bit and 250MB for 32-bit systems.

### Examples

Remove a file:
```
kure file rm example
```

Remove a directory:
``` 
kure file rm books/
```

Remove multiple files:
```
kure file rm Sample Sample2 Sample3
```
