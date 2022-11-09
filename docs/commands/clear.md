## Use

`kure clear  [-c clipboard] [-H history] [-t terminal]`

## Description

Clear clipboard, terminal screen or history.
		
Using the command without passing any flags clears the clipboard and the terminal screen.

## Flags

| Name | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| clipboard | c | bool | false | Clear clipboard |
| history | H | bool | false | Remove kure commands from terminal history |
| terminal | t | bool | false | Clear terminal screen |

## Examples

Clear terminal and clipboard:
```
kure clear
```

Clear clipboard:
```
kure clear -c
```

Clear terminal screen:
```
kure clear -t
```

Clear kure commands from terminal history:
```
kure clear -H
```