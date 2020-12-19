## Use

`kure file add <name> [-b buffer] [-i ignore] [-p path] [-s semaphore]`

*Aliases*: add, a.

## Description

Add files to the database.

Path to a file must include its extension (in case it has).

The user can specify a path to a folder as well, on this occasion, Kure will iterate over all the files in the folder and potential subfolders (if the -i flag is false) and store them into the database with the name "name/subfolders/filename".

Default behavior in case the buffer flag is not used:
   • file <= 1GB: read the entire file directly to memory.
   • file > 1GB: use a 64MB buffer.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |                     Description                   |
|-----------|-----------|---------------|---------------|---------------------------------------------------|
| buffer    | b         | uint64        | 0             | Buffer size when reading files                    |
| ignore    | i         | bool          | false         | Ignore subfolders                                 | 
| path      | p         | string        | ""            | Path to the file/folder                           |
| semaphore | s         | uint          | 1             | Maximum number of goroutines running concurrently |

The **buffer** flag is especially useful to avoid exhausting system's memory limit when adding a large file to the database (this could also improve performance if used correctly).

#### Goroutines

Goroutines can be thought of as light weight threads managed by the Go runtime. There might be only one thread in a program with thousands of Goroutines.

The cost of creating a Goroutine is tiny when compared to a thread, while the minimum stack size is defined as 2048 bytes, the Go runtime does also not allow goroutines to exceed a maximum stack size; this maximum depends on the architecture and is 1 GB for 64-bit and 250MB for 32-bit systems.

### Examples

Add a new file:
```
kure file add -p path/to/file
```

Add a folder and all its subfolders, limiting goroutine number to 40:
```
kure file add -p path/to/folder -s 40
```

Add files from a folder, ignoring subfolders and using a 4096 bytes buffer:
```
kure file add -p path/to/folder -i -b 4096
```