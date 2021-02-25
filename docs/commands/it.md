## Use

`kure it <command|flags|name>`

## Description

Interactive prompt.		
This command behaves depending on the arguments received, it requests the missing information.

|       Given       |    Requests       |
|-------------------|-------------------|
| command           | flags and name    |
| command and flags | name              |
| name              | command and flags |

## Flags 

No flags.

### Examples

No arguments:
```
kure it
```

Command without flags:
```
kure it ls
```

Command with flags:
```
kure it ls -s -q
```

Only the name:
```
kure sample
```
