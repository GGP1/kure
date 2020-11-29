## Use

`kure file add <name> [-i ignore] [-p path] [-s semaphore]`

*Aliases*: add, a.

## Description

Add files to the database.

The user can specify either a path to a file or to a folder, in case it points to a folder, Kure will iterate over all the files in the folder (ignoring sub folders) and store them into the database with the name: "(name)-(file number)".

## Flags 

|  Name     | Shorthand |     Type      |    Default    |                                       Usage                                            |
|-----------|-----------|---------------|---------------|----------------------------------------------------------------------------------------|
| buffer    | b         | uint64        | 0             | Buffer size when reading files (by default it reads the entire file directly to memory)|
| ignore    | i         | bool          | false         | Ignore subfolders                                                                      | 
| path      | p         | string        | ""            | Path to the file/folder                                                                |
| semaphore | s         | uint          | 1             | Maximum number of goroutines running concurrently                                      |

The **buffer** flag is especially useful to avoid exhausting system's memory limit when adding a large file to the database (this could also improve performance if used correctly).

#### Goroutines

Goroutines can be thought of as light weight threads managed by the Go runtime. There might be only one thread in a program with thousands of Goroutines.

The cost of creating a Goroutine is tiny when compared to a thread, while the minimum stack size is defined as 2048 bytes, the Go runtime does also not allow goroutines to exceed a maximum stack size; this maximum depends on the architecture and is 1 GB for 64-bit and 250MB for 32-bit systems.

### Examples

Add a file:
```
kure file add example -p path/to/file
```

Add all the files into a folder:
```
kure file add group -p path/to/folder
```