## Use

`kure file ls <name> [-f filter]`

## Description

List files.

## Flags

|  Name     | Shorthand |     Type      |    Default    |       Usage        |
|-----------|-----------|---------------|---------------|--------------------|
| filter    | f         | bool          | false         | Filter files       |

### Example

List passport file:
```
kure file ls passport
```

Filter among files:
```
kure file ls book -f
```

List all the files:
```
kure file ls
```