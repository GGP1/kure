## Use

`card`

## Description

Card operations.

## Flags

No flags.

## Subcommands

### Use

`add <name>`

### Description

Add a card to the database.

### Flags

No flags.

#### Examples

Add a card:
```
kure card add Sample
```

### Use 

`copy <name> [-t timeout] [field]`

### Description

Copy card number or cvc.

### Flags

|  Name     |  Shorthand    |     Type      |    Default    |                     Usage                     |
|-----------|---------------|---------------|---------------|-----------------------------------------------|
| timeout   | t             | duration      | 0             | Set a time until the clipboard is cleaned     |
| field     | f             | string        | "number"      | Set which card field to copy                  |

#### Examples

Copy card number and clean after 15 minutes:
```
kure card copy Sample -t 15m
```

Copy card CVC:
```
kure card copy Sample --field cvc
```

### Use 

`delete <name>`

### Description

Delete a card from the database.

### Flags

No flags.

#### Examples

Delete card:
```
kure card delete Sample
```

### Use 

`list <name>`

### Description

List a card or all the cards from the database.

### Flags

No flags.

#### Examples

List a specific card:
```
kure card list Sample
```

List all cards;
```
kure card list
```
