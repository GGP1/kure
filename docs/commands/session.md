## Use

`kure session [-p prefix] [-t timeout]`

## Description

Sessions are used for doing multiple operations by providing the master password once, it's encrypted and stored inside a locked buffer, decrypted when needed and destroyed right after it.

Scripts can be created in the configuration file and executed inside sessions by using their aliases and, optionally, passing arguments. They can be composed of *kure* and *session* commands but not other scripts.

> Adding scripts inside a session will require to restart it to take effect as they are loaded on the command initialization and not before every command.

Once into a session:
- use "&&" to execute a commands sequence.
- it's optional to use the word "kure" to run a command.

Session commands:
- block - block execution (to be manually unlocked).
- exit|quit - close the session (Ctrl+C is also an option).
- pwd - show current directory.
- timeout - show the session time left.
- timer - run a timer showing the session time left.
- ttadd [duration] - increase/decrease timeout.
- ttset [duration] - set a new timeout.
- sleep [duration] - sleep for x time.

## Flags

| Name | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| prefix | p | string | kure:~ $ | Text that precedes your commands |
| timeout | t | duration | 0s | Session timeout |

### Examples

Run a session without timeout and using "$" as the prefix:
```
kure session -p $
```

Run a session for 1 hour:
```
kure session -t 1h
```