## Use

`kure edit <name> [-i it]`

## Description

Edit an entry. 

If the name is edited, kure will remove the entry with the old name and create one with the new name.

**Caution**: when using a text editor the content of the entry is written in plaintext to a temporary file, although the file has a random name and it's erased right after the first save, this isn't secure enough.

Command procedure when using a text editor:
1. Create a temporary file and write the entry content encoded with JSON to it.
2. Execute the text editor to edit it.
3. Wait for it to be saved.
4. Read its content and update the entry.
5. Overwrite the file with random bytes and delete.

Tips:
- Use '\n' to add new lines.
- Some text editors will require to exit to modify the file.

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

|  Name     | Shorthand |     Type      |    Default    |     Description      |
|-----------|-----------|---------------|---------------|----------------------|
| it        | i         | bool          | false         | Use a text editor    |

### Examples

Edit entry using the standard input:
```
kure edit Sample 
```

Edit entry using a text editor:
```
kure edit Sample -i
```
