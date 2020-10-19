## Use

`file`

## Description

Save the encrypted bytes of the file in the database, delete or list them creating a file whenever you want. The user must specify the complete path, including the file type, for example: path/to/file/sample.png

In case you don't specify where to list the file, it will be created in your current directory, for listing all the files kure will
create a folder with the name "kure_files" in your directory and save all the files there.

## Flags 

No flags.

### Subcommands

kure file **add**: Add files to the database.

kure file **create**: Create one, many or all the files in the database. In case a path is passed, kure will create any missing folders for you.

kure file **delete**: Delete a file from the database.

kure file **list**: Display information about each file in the bucket.

