## Use

`kure config create [-p path]`

## Description

Create a configuration file. Kure will create the file for you if necessary.

### Formats supported

- YAML
- JSON
- TOML

#### Text editors commands
*Editor*: *command*
```
Vim: vim
Neovim: nvim
Emacs: emacs
Nano: nano
Visual Studio Code: code
Sublime Text: subl
Atom: atom
Coda: coda
Notepad: notepad
Notepad++: notepad++
...
```

## Flags 

|  Name     | Shorthand |     Type      |    Default    |      Description      |
|-----------|-----------|---------------|---------------|-----------------------|
| path      | p         | string        | ""            | Destination file path |

### Examples

Create a configuration file:
```
kure config create -p path/to/file
```