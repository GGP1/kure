## Use

`kure card edit <name> [-i it]`

## Description

Edit a card.

If the name is edited, Kure will remove the card with the old name and create one with the new name.

**Caution**: when using a text editor the content of the card is written in plaintext to a temporary file, although the file has a random name and it's erased right after the first save, this isn't secure enough.

Command procedure when using a text editor:
1. Create a temporary file and write the card content encoded with JSON to it.
2. Execute the text editor to edit it.
3. Wait for it to be saved.
4. Read content its and update the card.
5. Overwrite the file with random bytes and delete.

Tips:
- Use '\n' to add new lines.
- Some text editors will require to exit to modify the file.

#### Text editors commands
*Editor*: *value*
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

|  Name     | Shorthand |     Type      |    Default    |     Description    |
|-----------|-----------|---------------|---------------|--------------------|
| it        | i         | bool          | false         | Use text editor    |

### Examples

Edit card with standard input:
```
kure card edit Sample
```

Edit card with text editor:
```
kure card edit Sample -i
```