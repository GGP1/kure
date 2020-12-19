## Use

`kure clear [-b both] [-c clipboard] [-t terminal]`

## Description

Manually clean clipboard, terminal (and its history) or both of them. Kure clears all by default.

Windows users must clear the history manually with ALT+F7, executing "cmd" command or by re-opening the cmd (as it saves session history only).

## Flags 

|  Name     | Shorthand |     Type      |    Default    |           Description             |
|-----------|-----------|---------------|---------------|-----------------------------------|
| both      | b         | bool          | true          | Clear both clipboard and terminal |
| clipboard | c         | bool          | false         | Clear clipboard                   |
| terminal  | t         | bool          | false         | Clear terminal                    |

### Examples

> By default it clears the clipboard and the terminal

Clear both clipboard and terminal:
```
kure clear -b
```

Clear clipboard:
```
kure clear -c
```

Clear terminal:
```
kure clear -t
```
