## Use

`kure file ls <name> [-f filter]`

## Description

List files.

## Flags

|  Name     | Shorthand |     Type      |    Default    |      Description      |
|-----------|-----------|---------------|---------------|-----------------------|
| filter    | f         | bool          | false         | Filter files by name  |

### Example

List trip.txt file and copy its content to the clipboard:
```
kure file ls trip.txt
```

Filter among files:
```
kure file ls book -f
```

List all the files:
```
kure file ls
```