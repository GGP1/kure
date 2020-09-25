## Use

`wallet <name> [-a add | -c copy | -d delete | -l list | -v view] [-t timeout]`

## Description

Wallets operations.

## Flags 
|  Name     |  Shorthand    |     Type      |    Default    |            Usage               |
|-----------|---------------|---------------|---------------|--------------------------------|
| add       | a             | bool          | false         | Add a wallet                   |
| copy      | c             | bool          | false         | Copy wallet number             |
| delete    | d             | bool          | false         | Delete a wallet                |
| list      | l             | bool          | true          | List wallet/wallets            |
| view      | v             | bool          | false         | View wallets                   |
| timeout   | t             | duration      | 0             | Clipboard cleaning timeout     |

### Examples

Add a wallet:
```
kure wallet -a 
```

Copy wallet number:
```
kure wallet -c -t 15m
```

Delete wallet:
```
kure wallet -d
```

List a specific wallet:
```
kure wallet test -l  
```

List all wallets;
```
kure wallet -l
```
