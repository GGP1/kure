## Use

![kure copy](https://user-images.githubusercontent.com/51374959/98058727-2991a980-1e24-11eb-8371-41d5838748e5.png)

## Description

Copy entry credentials to clipboard.

## Flags

|  Name     |  Shorthand    |     Type      |    Default    |            Usage             |
|-----------|---------------|---------------|---------------|------------------------------|
| timeout   | t             | duration      | 0             | Clipboard cleaning timeout   |
| username  | u             | bool          | false         | Copy entry username          |

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

Copy password and clean clipboard after 5 minutes:
```
kure copy Github -t 5m
```
