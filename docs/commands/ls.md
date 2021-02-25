## Use

`kure ls <name> [-f filter] [-q qr] [-s show]`

*Aliases*: ls, entries.

## Description

List entries.

> Listing all the entries does not check for expired entries, this decision was taken to prevent high loads when the number of entries is elevated. Listing a single entry does notifies if it is expired.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |                                  Description                                         |
|-----------|-----------|---------------|---------------|--------------------------------------------------------------------------------------|
| filter    | f         | bool          | false         | Filter entries                                                                       |
| qr        | q         | bool          | false         | Show the password QR code on the terminal (non-available when listing all entries)   |
| show      | s         | bool          | false         | Show entry password                                                                  |

### Examples

List an entry:
```
kure ls Sample
```

List one and show sensitive information:
```
kure ls Sample -s
```

Filter:
```
kure ls Sample -f
```

List all entries:
```
kure ls
```