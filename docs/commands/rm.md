## Use

`kure rm <name> [-d dir]`

## Description

Remove an entry or a directory.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |      Description      |
|-----------|-----------|---------------|---------------|-----------------------|
| dir       | d         | bool          | false         | Remove a directory    |

### Examples

Remove:
```
kure rm Sample
```

Remove a directory:
```
kure rm dirName -d
```