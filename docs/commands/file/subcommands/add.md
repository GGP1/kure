## Use

`kure file add <name> [-i ignore] [-n note] [-p path] [-s semaphore]`

*Aliases*: new.

## Description

Add files to the database. As they are stored in a database, the whole file is read into memory, please have this into account when adding new ones.

Path to a file must include its extension (in case it has).

The user can specify a path to a folder as well, on this occasion, Kure will iterate over all the files in the folder and potential subfolders (if the -i flag is false) and store them into the database with the name "name/subfolders/filename". Empty folders will be skipped.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |                     Description                   |
|-----------|-----------|---------------|---------------|---------------------------------------------------|
| ignore    | i         | bool          | false         | Ignore subfolders                                 | 
| note      | n         | bool          | false         | Add a note                                        | 
| path      | p         | string        | ""            | Path to the file/folder                           |
| semaphore | s         | uint          | 50            | Maximum number of goroutines running concurrently |

#### Goroutines

Goroutines can be thought of as light weight threads managed by the Go runtime. There might be only one thread in a program with thousands of Goroutines.

The cost of creating a Goroutine is tiny when compared to a thread, while the minimum stack size is defined as 2048 bytes, the Go runtime does also not allow goroutines to exceed a maximum stack size; this maximum depends on the architecture and is 1 GB for 64-bit and 250MB for 32-bit systems.

### Examples

Add a new file:
```
kure file add Sample -p path/to/file
```

Add a note:
```
kure file add Sample -n
```

Add a folder and all its subfolders, limiting goroutine number to 40:
```
kure file add Sample -p path/to/folder -s 40
```

Add files from a folder, ignoring subfolders:
```
kure file add Sample -p path/to/folder -i 
```