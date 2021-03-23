## Use

`kure 2fa add <name> [-d digits] [-u url]`

## Description

Add a two-factor authentication code.

- **Using a setup key**: services tipically show hyperlinked text like "Enter manually" or "Enter this text code", copy the hexadecimal code given and submit it when requested.

- **Using a URL**: extract the URL encoded in the QR code given and submit it when requested. Format: otpauth://totp/{service}:{account}?secret={secret}.

## Flags

| Name | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| digits | d | int32 | 6 | TOTP length {6|7|8} |
| url | u | bool | false | Add using a URL |

### Examples

Add with setup key:
```
kure 2fa add Sample
```

Add with URL:
```
kure 2fa add -u
```