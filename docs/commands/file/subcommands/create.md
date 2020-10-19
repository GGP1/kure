## Use

`create <name> [-p path]`

## Description

Create one, many or all the files in the database. In case a path is passed, kure will create any missing folders for you.

## Flags 

|  Name     |  Shorthand    |     Type      |    Default    |                             Usage                                 |
|-----------|---------------|---------------|---------------|-------------------------------------------------------------------|
| many      | m             | []string      | nil           | A list of files to create                                         |
| path      | p             | string        | ""            | Destination path (this is where the file will be created)         |
| overwrite | o             | bool          | false         | Set to overwrite files if they already exist in the path provided |

## Examples

Create a specific file overwriting if exists:
```
kure file create example -p path/to/destination -o
```

Create many files:
```
kure file create -m file1,file2
```

Create all the files in the database:
```
kure file create
```