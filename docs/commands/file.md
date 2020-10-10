## Use

`file`

## Description

Save the encrypted bytes of the file in the database, delete or list them creating a file whenever you want. The user must specify the complete path, including the file type, for example: path/to/file/sample.png

In case you don't specify where to list the file, it will be created in your current directory, for listing all the files kure will
create a folder with the name "kure_files" in your directory and save all the files there.

## Flags 

No flags.

## Subcommands

### Use

`add <name> [-p path]`

### Description

Add a file to the database.

### Flags 

|  Name     |  Shorthand    |     Type      |    Default    |            Usage             |
|-----------|---------------|---------------|---------------|------------------------------|
| path      | p             | string        | ""            | Path to the file file       |

#### Examples

Add a file:
```
kure file add example -p path/to/file
```

### Use

`delete <name>`

### Description

Delete a file from the database.

### Flags 

No flags.

#### Examples


Delete a file:
```
kure file delete example
```

### Use

`info <name>`

### Description

Display information about each file in the bucket.

### Flags

No flags.

#### Example

Passport file info:
```
kure file info passport
```

### Use

`list <name> [-p path]`

### Description

List a file or all the files from the database.

### Flags 

|  Name     |  Shorthand    |     Type      |    Default    |                         Usage                             |
|-----------|---------------|---------------|---------------|-----------------------------------------------------------|
| path      | p             | string        | ""            | Destination path (this is where the file will be created) |

#### Examples

List a specific file:
```
kure file list example -p path/to/destination
```

List all files;
```
kure file list
```