## Use 

![kure wallet ls](https://user-images.githubusercontent.com/51374959/98058971-be94a280-1e24-11eb-8450-5c49fd0e9dee.png)

## Description

List wallets.

## Flags

|  Name     | Shorthand |     Type      |    Default    |       Usage        |
|-----------|-----------|---------------|---------------|--------------------|
| filter    | f         | bool          | false         | Filter wallets     |

### Examples

List a specific wallet:
```
kure wallet ls Satoshi
```

Filter among wallets:
```
kure wallet ls btc -f
```

List all wallets;
```
kure wallet ls
```
