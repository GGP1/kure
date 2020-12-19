## Use

`kure file cat <name> [-c copy]`

## Description

Read file and print to standard output

## Flags 

|  Name     | Shorthand |     Type      |    Default    |              Description              |
|-----------|-----------|---------------|---------------|---------------------------------------|
| copy      | c         | bool          | false         | Copy file content to the clipboard    |

### Examples

Read one file:
```
kure cat fileName
```

Read one file and copy content to the clipboard:
```
kure cat fileName -c
```

Read multiple files:
```
kure cat file1 file2 file3
```