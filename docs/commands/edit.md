## Use

`kure edit <name> [-n name]`

## Description

Edit entry fields.
	
"-" = Clear field.

"" (nothing) = Do not modify field.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |           Usage            |
|-----------|-----------|---------------|---------------|----------------------------|
| name      | n         | bool          | false         | Edit entry name as well    |

### Expires

Valid time formats are: 

• ISO: 2006/01/02 or 2006-01-02.

• US: 02/01/2006 or 02-01-2006.

"never" and "" will be considered as if the entry never expires.

### Examples

Edit an entry:
```
kure edit Twitter
```

Edit an entry password:
```
kure edit Google -p
```