## Use

![kure file rm](https://user-images.githubusercontent.com/51374959/98058775-4af29580-1e24-11eb-9ccc-d40c3a5c483f.png)

## Description

Delete files from the database.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |                       Usage                             |
|-----------|-----------|---------------|---------------|---------------------------------------------------------|
| dir       | d         | bool          | false         | Remove a directory and all the files stored in it       |

### Examples

Remove a file:
```
kure file rm example
```

Remove a directory and all the files in it:
``` 
kure file rm books -d
```