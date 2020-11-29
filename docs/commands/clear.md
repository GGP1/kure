## Use

`kure clear [-b both] [-c clipboard] [-t terminal]`

## Description

Manually clean clipboard, terminal (and its history) or both of them.
Windows users must clear the history manually with ALT+F7, executing "cmd" command or by re-opening the cmd (as it saves session history only).

## Flags 

|  Name     | Shorthand |     Type      |    Default    |            Usage                  |
|-----------|-----------|---------------|---------------|-----------------------------------|
| both      | b         | bool          | true          | Clear both clipboard and terminal |
| clipboard | c         | bool          | false         | Clear clipboard                   |
| terminal  | t         | bool          | false         | Clear terminal                    |

### Examples

Clear both clipboard and terminal:
```
kure clear
```
(by default it will clear both clipboard (its history also) and terminal)

Clear clipboard:
```
kure clear -c
```

Clear terminal:
```
kure clear -t
```
