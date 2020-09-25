## Use

`clear [-b both] [-c clipboard] [-t terminal]`

## Description

Manually clean the clipboard, terminal or both of them.

## Flags 
|  Name     |  Shorthand    |     Type      |    Default    |            Usage                  |
|-----------|---------------|---------------|---------------|-----------------------------------|
| both      | b             | bool          | true          | Clear both clipboard and terminal |
| clipboard | c             | bool          | false         | Clear clipboard                   |
| terminal  | t             | bool          | false         | Clear terminal                    |

### Examples

Clear both clipboard and terminal:
```
kure clear
```
(including -b is optional as it is set to true by default)

Clear clipboard:
```
kure clear -c
```

Clear terminal:
```
kure clear -t
```
