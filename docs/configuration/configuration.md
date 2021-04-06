## Configuration

By default Kure will read the file at `$HOME/.kure/kure.yaml` or the one specified in the `KURE_CONFIG` environment variable in case it is set. To change the file used, simply change the environment variable.

Paths inside it **MUST** be **absolute**.

*Formats supported*: JSON, TOML, YAML. [Samples](https://github.com/GGP1/kure/tree/master/docs/configuration/samples).

#### Helpful commands
1.  [`kure config`](https://github.com/GGP1/kure/tree/master/docs/commands/config/config.md) -> Read current file
1.  [`kure config create`](https://github.com/GGP1/kure/tree/master/docs/commands/config/subcommands/create/create.md) -> Create a new file
1.  [`kure config edit`](https://github.com/GGP1/kure/tree/master/docs/commands/config/subcommands/edit/edit.md) -> Edit current file

### Keys

- [Clipboard](#clipboard)
  - [Timeout](#timeout)
- [Database](#database)
  - [Path](#path)
- [Editor](#editor)
- [Keyfile](#keyfile)
  - [Path](#path)
- [Session](#session)
  - [Prefix](#prefix)
  - [Scripts](#scripts)
  - [Timeout](#timeoutt)

---

### Clipboard
#### Timeout

Time until the clipboard is cleared after a record has been copied to it.
Set to "0s" or leave blank for no timeout.

---

### Database
#### Path

> Must be absolute.

Path to the database file (if it doesn't exist, it will be created).

---

### Editor

The command of the editor you would like to use. If no editor is set in the configuration file, Kure will look for it in the `$EDITOR` and `$VISUAL` environment variables, if still nothing is found, it will try using vim by default.

---

### Keyfile
#### Path

> Must be absolute.

The path to the key file may be specified or not, in case it's not, the user will be asked for it everytime he wants to access the database, in the other case the user has to input the password only.

---

### Session
#### Prefix

Text that precedes your commands.

#### Scripts

Scripts can be used to run a sequence of commands inside sessions. Each one of them has an alias and may contain one-based indexing arguments ($1, $2, ..., $n) to be replaced by the arguments passed when executing the script. For example, having the script `list: ls $1 -s` we execute it by typing `list sample`, that is `<alias> <$1>`.

> Aliases must not contain spaces and arguments containing spaces must be enclosed by double quotes.

#### Timeout

Time until the session is closed.
Set to "0s" or leave blank for no timeout.ç