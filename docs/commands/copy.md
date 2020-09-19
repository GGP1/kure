## Use

`copy <title> [-t terminal]`

## Description

Copy password to clipboard for t time (never by default).

## Flags 
```
|  Name     |  Shorthand    |     Type      |    Default    |            Usage             |
|-----------|---------------|---------------|---------------|------------------------------|
| timeout   | t             | duration      | 0             | Clipboard cleaning timeout   |
```

### Examples

Simple copy:
```
kure copy Twitter
```

Copy with timeout:
```
kure copy Twitter -t 20m
```
