## Use

`kure file edit <name> [-e editor]`

## Description

Edit a file.

Command procedure:
1. Create a file with the content of the stored file.
2. Execute the text editor to edit it.
3. Wait for it to be saved.
4. Read its content and update kure's file.
5. Overwrite the initially created file with random bytes and delete.

Tips:
- Some editors will require to exit to modify the file.
- Use an image editor command to edit images.

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

#### Image editors commands
*Editor*: *command*
```
GIMP: gimp
Paint: mspaint
Krita: krita
...
```

## Flags

|  Name     | Shorthand |     Type      |    Default    |      Description     |
|-----------|-----------|---------------|---------------|----------------------|
| editor    | e         | string        | ""            | File editor command  |

### Examples

Edit file:
```
kure file edit Sample -e vim
```