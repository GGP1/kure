## Use

`kure config [-c create] [-p path]`

*Aliases*: config, cfg.

## Description

Read or create the configuration file.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |            Usage             |
|-----------|-----------|---------------|---------------|------------------------------|
| create    | c         | bool          | false         | Create a config file         |
| path      | p         | string        | ""            | Config file path             |

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