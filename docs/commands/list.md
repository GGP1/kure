## Use

`list <title> [-H hide]`

## Description

List entry/entries.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |          Usage           |
|-----------|-----------|---------------|---------------|--------------------------|
| hide      | H         | bool          | false         | Hide entries passwords   |

### Examples

List an entry:
```
kure list Binance
```

List an entry hiding the password:
```
kure list Youtube -H
```

List all entries:
```
kure list
```