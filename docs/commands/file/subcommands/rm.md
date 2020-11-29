## Use

`kure file rm <name> [-d dir] [-s semaphore]`

## Description

Delete files from the database.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |                       Usage                           |
|-----------|-----------|---------------|---------------|-------------------------------------------------------|
| dir       | d         | bool          | false         | Remove a directory and all the files stored in it     |
| semaphore | s         | uint          | 1             | Maximum number of goroutines running concurrently     |

#### Goroutines

Goroutines can be thought of as light weight threads managed by the Go runtime. There might be only one thread in a program with thousands of Goroutines.

The cost of creating a Goroutine is tiny when compared to a thread, while the minimum stack size is defined as 2048 bytes, the Go runtime does also not allow goroutines to exceed a maximum stack size; this maximum depends on the architecture and is 1 GB for 64-bit and 250MB for 32-bit systems.

### Examples

Remove a file:
```
kure file rm example
```

Remove a directory using a maximum of 20 goroutines:
``` 
kure file rm books -d -s 20
```