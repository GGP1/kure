## Use

`kure file cat <name> [-c copy]`

## Description

Read file and write to standard output

## Flags 

|  Name     | Shorthand |     Type      |    Default    |              Description              |
|-----------|-----------|---------------|---------------|---------------------------------------|
| copy      | c         | bool          | false         | Copy file content to the clipboard    |

### Examples

Write one file:
```
kure cat fileName
```

Write one file and copy content to the clipboard:
```
kure cat fileName -c
```

Write multiple files:
```
kure cat file1 file2 file3
```