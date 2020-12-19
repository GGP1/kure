# Commands list

#### Backup
```
backup [http] [port] [path]
```

#### Card
```
card
  copy <name> [-c cvc] [-t timeout]
  create <name>
  ls <name> [-f filter] [-H hide]
  rm <name> [-d dir]
```

#### Clear
```
clear [-b both] [-c clipboard] [-t terminal]
```

#### Config
```
config [-c create] [-p path]
  test
```

#### Copy
```
copy <name> [-t timeout] [-u username]
```

#### Create
```
create <name> [-c custom] [-l length] [-f format] [-i include] [-e exclude] [-r repeat]
    phrase <name> [-l length] [-s separator] [-i include] [-e exclude] [list]
```

#### Edit 
```
edit <name> [-n name]
```

#### Export
```
export <manager-name> [-p path]
```

#### File
```
file
  add <name> [-b buffer] [-i ignore] [-p path] [-s semaphore]
  cat <name> [-c copy]
  ls <name> [-f filter]
  rm <name> [-d dir]
  rename <oldName> <newName>
  touch <name> [-o overwrite] [-p path]
```

#### Gen
```
gen [-c copy] [-l length] [-f format] [-i include] [-e exclude] [-r repeat] [-q qr]
    phrase [-c copy] [-l length] [-s separator] [-i include] [-e exclude] [list] [-q qr]
```

#### Help
```
help
```

#### Import
```
import <manager-name> [-p path]
```

#### Ls
```
ls <name> [-f filter] [-H hide] [-q qr]
```

#### Note
```
note
  copy <name> [-t timeout]
  create <name>
  ls <name> [-f filter]
  rm <name> [-d dir]
```

#### Restore
```
restore
  argon2 [-i iterations] [-m memory] [-t threads]
  password
```

#### Rm
```
rm <name> [-d dir]
```

#### Session
```
session [-p prefix] [-t timeout]
```

#### Stats 
```
stats
```