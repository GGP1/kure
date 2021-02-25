## Use

`kure 2fa add <name> [-d digits]`

## Description

Add a two-factor authentication code. The name must be one already used by an entry.

Services tipically show an hyperlinked "Enter manually", "Enter this text code" or similar messages, copy the hexadecimal code given and submit it when requested by Kure. After this, your entry will have a synchronized token with the service.

## Flags

|  Name     |     Type      |    Default    |            Description            |
|-----------|---------------|---------------|-----------------------------------|
| digits    | int           | 6             | TOTP length                       |

### Examples

Add a 2FA code to Sample (Sample must be an already created entry):
```
kure 2fa add Sample
```