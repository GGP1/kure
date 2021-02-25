## Use

`kure clear [-c clipboard] [-t terminal]`

## Description

Manually clear the clipboard, terminal or both of them. Kure clears all by default.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |           Description             |
|-----------|-----------|---------------|---------------|-----------------------------------|
| clipboard | c         | bool          | false         | Clear clipboard                   |
| terminal  | t         | bool          | false         | Clear terminal                    |

### Examples

> By default it clears the clipboard and the terminal

Clear both clipboard and terminal:
```
kure clear
```

Clear clipboard:
```
kure clear -c
```

Clear terminal:
```
kure clear -t
```
