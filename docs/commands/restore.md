## Use 

`kure restore`

## Description

Restore the database using new credentials.

Overwrite the registered credentials and re-encrypt every record with the new ones.

Warning: all the records will be stored in memory during the process, restoring a big set of them can cause an OOM error. In these cases it's preferred to create a database with new credentials and use kure commands to write and read the data from the filesystem.

## Flags

No flags.
