## Use

`add <name> [-p path]`

## Description

Add files to the database.

The user can specify either a path to a file or to a folder, in case it points to a folder, Kure will iterate over all the files in the folder (ignoring sub folders) and store them into the database with the name: "(name)-(file number)".

## Flags 

|  Name     |  Shorthand    |     Type      |    Default    |            Usage             |
|-----------|---------------|---------------|---------------|------------------------------|
| path      | p             | string        | ""            | Path to the file file       |

### Examples

Add a file:
```
kure file add example -p path/to/file
```
