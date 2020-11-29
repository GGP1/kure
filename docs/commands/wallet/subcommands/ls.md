## Use 

`kure wallet ls <name> [-f filter] [-H hide]`

## Description

List wallets.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                 Usage                     |
|-----------|-----------|---------------|---------------|-------------------------------------------|
| filter    | f         | bool          | false         | Filter wallets                            |
| hide      | H         | bool          | false         | Hide wallet seed phrase and private key   |


### Examples

List a specific wallet hiding critical information:
```
kure wallet ls Satoshi -H
```

Filter among wallets:
```
kure wallet ls btc -f
```

List all wallets;
```
kure wallet ls
```
