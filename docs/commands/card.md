## Use

`card <name> [-a add | -c copy | -d delete | -l list | -v view] [-t timeout] [-f field]`

## Description

Cards operations.

## Flags 
```
|  Name     |  Shorthand    |     Type      |    Default    |            Usage                  |
|-----------|---------------|---------------|---------------|-----------------------------------|
| add       | a             | bool          | false         | Add a card                        |
| copy      | c             | bool          | false         | Copy card number                  |
| delete    | d             | bool          | false         | Delete a card                     |
| list      | l             | bool          | true          | List card/cards                   |
| view      | v             | bool          | false         | View wallets                      |
| timeout   | t             | duration      | 0             | Clipboard cleaning timeout        |
| field     | f             | string        | "number"      | Choose which card field to copy   |
```

### Examples

Add a card:
```
kure card -a 
```

Copy card number:
```
kure card -c -t 15m
```

Copy card CVC:
```
kure card -c -t 15m -f cvc
```

Delete card:
```
kure card -d
```

List a specific card:
```
kure card test -l  
```

List all cards;
```
kure card -l
```
