## Use

`kure import <manager-name> [-p path]`

## Description

Import entries from other password managers. Format: CSV.

Supported:
    • Bitwarden
    • Keepass
    • Lastpass
    • 1Password

## Flags

|  Name     | Shorthand |     Type      |    Default    |     Description      |
|-----------|-----------|---------------|---------------|----------------------|
| path      | p         | string        | ""            | Path to csv file     |

### Examples

Import:
```
kure import <manager-name> -p path/to/file
```