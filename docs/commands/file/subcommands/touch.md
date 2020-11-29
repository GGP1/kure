## Use

`kure file touch <name> [-m multiple] [-o overwrite] [-p path]`

*Aliases*: touch, th, t.

## Description

Create one, multiple or all the files in the database. In case a path is passed, kure will create any missing folders for you.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                             Usage                                 |
|-----------|-----------|---------------|---------------|-------------------------------------------------------------------|
| overwrite | o         | bool          | false         | Set to overwrite files if they already exist in the path provided |
| path      | p         | string        | ""            | Destination path (this is where the file will be created)         |
| multiple  | m         | []string      | nil           | A list of files to create                                         |

## Examples

Create a specific file overwriting if exists:
```
kure file create example -p path/to/destination -o
```

Create multiple files:
```
kure file create -m file1,file2
```

Create all files (tree recreation):
```
kure file create -p path/to/folder/new/new2
```

Kure will recreate the file tree inside new2 folder which is inside path/to/folder/new.

If the user doesn't include a path, Kure will create them inside the directory the user is located.