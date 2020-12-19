## Use

`kure note rm <name> [-d dir]`

## Description

Remove a note or directory.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |      Description      |
|-----------|-----------|---------------|---------------|-----------------------|
| dir       | d         | bool          | false         | Remove a directory    |

### Examples

Remove note:
```
kure note rm Sample
```

Remove a directory:
```
kure note rm dirName -d
```