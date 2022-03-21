## Use

`kure file touch <name> [-o overwrite] [-p path]`

*Aliases*: th.

## Description

Create one, multiple, all the files or an specific directory.

For creating an specific file the extension must be included in the arguments, if not, Kure will consider that the user is trying to create a directory and it will search for it.

In case a path is passed, Kure will create any missing folders for you.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                          Description                              |
|-----------|-----------|---------------|---------------|-------------------------------------------------------------------|
| overwrite | o         | bool          | false         | Set to overwrite files if they already exist in the path provided |
| path      | p         | string        | ""            | Destination folder path                                           |

## Examples

Create a specific file overwriting if exists:
```
kure file touch example -p path/to/folder -o
```

Create multiple files and a directory in the current directory:
```
kure file touch file1,directory,file3
```

Create all files (tree recreation):
```
kure file touch -p path/to/folder/new/new2
```

Kure will recreate the file tree inside new2 folder which is inside path/to/folder/new.

If the user doesn't include a path, Kure will create them inside the directory the user is located.