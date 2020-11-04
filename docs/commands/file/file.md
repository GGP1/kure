## Use

![kure file](https://user-images.githubusercontent.com/51374959/98058779-4b8b2c00-1e24-11eb-8d48-eebc3e177e23.png)

## Description

Save the encrypted bytes of the file in the database, create it back whenever you want, list information or remove them. The user must specify the complete path, including the file type, for example: path/to/file/sample.png

In case you don't specify where to list the file, it will be created in your current directory, for listing all the files kure will
create a folder with the name "kure_files" in your directory and save all the files there.

## Flags 

No flags.

### Subcommands

- kure file **add**: Add files to the database.

- kure file **create**: Create one, many or all the files in the database. In case a path is passed, kure will create any missing folders for you.

- kure file **list**: Display information about each file in the bucket.

- kure file **rename**: Rename a file.

- kure file **rm**: Remove a file from the database.