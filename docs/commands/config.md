## Use

![kure config](https://user-images.githubusercontent.com/51374959/98058707-1ed71480-1e24-11eb-84da-ff8897b3146d.png)

## Description

Read or create the configuration file.

## Flags 

|  Name     |  Shorthand    |     Type      |    Default    |            Usage             |
|-----------|---------------|---------------|---------------|------------------------------|
| create    | c             | bool          | false         | Create a config file         |
| path      | p             | string        | ""            | Config file path             |

### Examples

Read configuration file:
```
kure config
```

Read specifying file path:
```
kure config -p path/to/file
```

Create configuration file:
```
kure config -c
```