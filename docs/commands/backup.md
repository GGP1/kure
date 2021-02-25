## Use

`kure backup [http] [path] [port]`

## Description

Create database backup.

## Flags

|  Name     |     Type      |    Default    |                  Description                   |
|-----------|---------------|---------------|------------------------------------------------|
| http      | bool          | false         | Serve the database file on a http server       |
| path      | string        | ""            | Backup file path                               |
| port      | uint16        | 8080          | Server port                                    |

### Examples

Create file backup:
```
kure backup --path path/to/file
```

Serve database on a local server:
```
kure backup --http --port 8080
```

Download database:
```
curl localhost:8080 > database_name
```