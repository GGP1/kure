## Use

![kure file ls](https://user-images.githubusercontent.com/51374959/98058770-4928d200-1e24-11eb-88cd-3f40b7c5d21f.png)

## Description

List files.

## Flags

|  Name     | Shorthand |     Type      |    Default    |       Usage        |
|-----------|-----------|---------------|---------------|--------------------|
| filter    | f         | bool          | false         | Filter files       |

### Example

List passport file:
```
kure file ls passport
```

Filter among files:
```
kure file ls book -f
```

List all the files:
```
kure file ls
```