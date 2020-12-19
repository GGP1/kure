## Use 

`kure note ls <name> [-f filter]`

## Description

List notes.

## Flags

|  Name     | Shorthand |     Type      |    Default    |    Description     |
|-----------|-----------|---------------|---------------|--------------------|
| filter    | f         | bool          | false         | Filter notes       |

### Examples

List a specific note:
```
kure note ls sample 
```

Filter among notes:
```
kure file ls ple -f
```

List all notes;
```
kure note ls
```