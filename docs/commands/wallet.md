## Use

`wallet`

## Description

Wallet operations.

## Flags 

No flags.

## Subcommands

### Use

`add <name>`

### Description

Add a wallet to the database.

### Flags

No flags.

#### Examples

Add a wallet:
```
kure wallet add Satoshi 
```

### Use 

`copy <name> [-t timeout]`

### Description

Copy wallet public key.

### Flags

|  Name     |  Shorthand    |     Type      |    Default    |                     Usage                     |
|-----------|---------------|---------------|---------------|-----------------------------------------------|
| timeout   | t             | duration      | 0             | Set a time until the clipboard is cleaned     |

#### Examples

Copy wallet number:
```
kure wallet copy Satoshi
```

### Use 

`delete <name>`

### Description

Delete a wallet from the database.

### Flags

No flags.

#### Exaples

Delete wallet:
```
kure wallet delete Satoshi
```

### Use 

`list <name>`

### Description

List a wallet or all the wallets from the database.

### Flags

No flags.

#### Examples

List a specific wallet:
```
kure wallet list Satoshi
```

List all wallets;
```
kure wallet list
```
