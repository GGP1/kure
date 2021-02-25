## Use

`kure copy <name> [-t timeout] [-u username]`

*Aliases*: copy, cp.

## Description

Copy entry credentials to clipboard.

## Flags

|  Name     | Shorthand |     Type      |    Default    |         Description          |
|-----------|-----------|---------------|---------------|------------------------------|
| timeout   | t         | time.Duration | 0             | Clipboard clearing timeout   |
| username  | u         | bool          | false         | Copy entry username          |

### Timeout units

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

### Examples

Copy password:
```
kure copy Github
```

Copy username:
```
kure copy Github -u
```

Copy password and clear clipboard after 5 minutes:
```
kure copy Github -t 5m
```
