## Use

`kure export <manager-name> [-p path]`

## Description

Export entries to other password managers. Format: CSV.

Supported:
    • Bitwarden
    • Keepass
    • Lastpass
    • 1Password

## Flags

|  Name     | Shorthand |     Type      |    Default    |       Description      |
|-----------|-----------|---------------|---------------|------------------------|
| path      | p         | string        | ""            | Destination file path  |

### Examples

Export:
```
kure export <manager-name> -p path/to/file
```