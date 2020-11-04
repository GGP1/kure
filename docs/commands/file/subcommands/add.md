## Use

![kure file add](https://user-images.githubusercontent.com/51374959/98058767-475f0e80-1e24-11eb-8b08-c6c744fe8e11.png)

## Description

Add files to the database.

The user can specify either a path to a file or to a folder, in case it points to a folder, Kure will iterate over all the files in the folder (ignoring sub folders) and store them into the database with the name: "(name)-(file number)".

## Flags 

|  Name     |  Shorthand    |     Type      |    Default    |                               Usage                                   |
|-----------|---------------|---------------|---------------|-----------------------------------------------------------------------|
| path      | p             | string        | ""            | Path to the file/folder                                               |
| semaphore | s             | uint          | 25            | Maximum number of goroutines to run when adding files to the database |
| ignore    | i             | bool          | false         | Ignore subfolders                                                     | 

### Examples

Add a file:
```
kure file add example -p path/to/file
```

Add all the files into a folder:
```
kure file add group -p path/to/folder
```
The first file will be stored as group-1, the second group-2 and so on.