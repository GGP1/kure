## Use

`kure file <subcommand>`

## Description

Save the encrypted bytes of the file in the database, create it back whenever you want, list information or remove them. The user must specify the complete path, including the file type, for example: path/to/file/sample.png

In case you don't specify where to list the file, it will be created in your current directory, for listing all the files kure will
create a folder with the name "kure_files" in your directory and save all the files there.

## Flags 

No flags.

### Subcommands

- kure file **add**: Add files to the database.

- kure file **list**: List files.

- kure file **rename**: Rename a file.

- kure file **rm**: Remove a file from the database.

- kure file **touch**: Create one, multiple or all the files in the database. In case a path is passed, kure will create any missing folders for you.