## Use

`backup [http | encrypt | decrypt] [port] [path]`

## Description

Create database backups, serve the database file on a local http server.

## Flags

|  Name     |     Type      |    Default    |                  Usage                         |
|-----------|---------------|---------------|------------------------------------------------|
| http      | bool          | false         | Serve the database file on a http server       |
| port      | uint16        | 4000          | Set server port                                |
| encrypt   | bool          | false         | Create encrypted backup                        |
| decrypt   | bool          | false         | Decrypt encrypted backup and read              |
| path      | string        | "./backup"    | Backup file path                               |

### Examples

Serve database on a server:
```
kure backup -http -p 4000
```

Download database:
```
curl localhost:4000 > kure.db
```

Encrypt file:
```
kure backup --encrypt --path path/to/file
```

Decrypt file:
```
kure backup --decrypt --path path/to/file
```