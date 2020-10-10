## Use

`copy <name> [-t timeout]`

## Description

Copy password to clipboard for t time (no timeout by default).

## Flags

|  Name     |  Shorthand    |     Type      |    Default    |            Usage             |
|-----------|---------------|---------------|---------------|------------------------------|
| timeout   | t             | duration      | 0             | Clipboard cleaning timeout   |

### Examples

Simple copy:
```
kure copy Github
```

Copy with and clean after 20 minutes:
```
kure copy Github -t 20m
```
