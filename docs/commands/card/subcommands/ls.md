## Use 

`kure card ls <name> [-f filter] [-H hide]`

## Description

List cards.

## Flags

|  Name     | Shorthand |     Type      |    Default    |    Description     |
|-----------|-----------|---------------|---------------|--------------------|
| filter    | f         | bool          | false         | Filter cards       |
| hide      | H         | bool          | false         | Hide card CVC      |

### Examples

List a specific card hiding CVC:
```
kure card ls sample -H
```

Filter among cards:
```
kure file ls ple -f
```

List all cards;
```
kure card ls
```