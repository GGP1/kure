## Use

![kure edit](https://user-images.githubusercontent.com/51374959/98058742-37472f00-1e24-11eb-8ba0-78d475686255.png)

## Description

Edit entry fields.
	
"-" = Clear field.
"" (nothing) = Do not modify field.

## Flags 

|  Name     |  Shorthand    |     Type      |    Default    |            Usage             |
|-----------|---------------|---------------|---------------|------------------------------|
| password  | p             | bool          | false         | Edit entry password (only)   |

### Examples

Edit an entry:
```
kure edit Twitter
```

Edit an entry password:
```
kure edit Google -p
```