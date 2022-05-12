## Use

`kure file edit <name>  [-e editor] [-l log]`

## Description

Edit a file.

Caution: a temporary file is created with a random name, it will be erased right after the first save but it could still be read by a malicious actor.
Notes:
    - Some editors flush the changes to the disk when closed, Kure won't notice any modifications until then.
    - Modifying the file with a different program will prevent Kure from erasing the file as its being blocked by another process.

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
| log | l | bool | false | Log the temporary file path and wait for modifications |

### Examples

Edit a file:
```
kure file edit Sample -e nvim
```

Write a file's content to a temporary file and log its path:
```
kure file edit Sample -l
```