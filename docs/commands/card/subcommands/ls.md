## Use 

`kure card ls <name> [-f filter] [-q qr] [-s show]`

## Description

List cards.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                 Description                   |
|-----------|-----------|---------------|---------------|-----------------------------------------------|
| filter    | f         | bool          | false         | Filter cards                                  |
| qr        | q         | bool          | false         | Display card number QR code on the terminal   |
| show      | s         | bool          | false         | Show card number and security code            |

### Examples

List a card showin sensitive information:
```
kure card ls Sample -s
```

Filter:
```
kure file ls Sample -f
```

List all cards;
```
kure card ls
```