## Use

`kure card rm <name> [-d dir]`

## Description

Remove a card or directory.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |      Description      |
|-----------|-----------|---------------|---------------|-----------------------|
| dir       | d         | bool          | false         | Remove a directory    |

### Examples

Remove card:
```
kure card rm Sample
```

Remove a directory:
```
kure card rm SampleDir -d
```