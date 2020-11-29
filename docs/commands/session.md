## Use

`kure session [-p prefix] [-t timeout]`

### Description

Sessions are used for doing multiple operations by providing the master password once, it's encrypted
and stored inside a locked buffer, decrypted when needed and destroyed right after it.

The user can set a timeout to automatically close the session after *X* amount of time. By default it never ends.

Once into the session:
• it's optional to use the word "kure" to run a command.
• type "timeout" to see the time left.
• type "exit" or press CTRL+C to quit.

### Flags

|  Name     | Shorthand |     Type      |    Default    |                      Usage                        |
|-----------|-----------|---------------|---------------|---------------------------------------------------|
| prefix    | p         | string        | "kure:~#"     | Customize the text that precedes your commands    |
| timeout   | t         | duration      | 0             | Session timeout. By default it never ends         |

### Timeout units

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

### Examples

Run a session for 1 hour
```
kure session -t 1h
```

Run a session without timeout and using "kure:~$" as the prefix
```
kure session -l kure:~$
```

Show the session time left (once into one)
```
timeout
```