## Use 

`kure restore argon2 [-i iterations] [-m memory] [-t threads]`

## Description

Re-encrypt all the information with new argon2 parameters.

## Flags

|  Name       | Shorthand |     Type      |       Default       |                 Description                   |
|-------------|-----------|---------------|---------------------|-----------------------------------------------|
| iterations  | i         | uint32        | 1                   | Number of passes over the memory              |
| memory      | m         | uint32        | 1048576             | Amount of memory allowed for argon2 to use    |
| threads     | t         | uint8         | N of CPUs available | Number of threads running in parallel         |

### Examples

Restore using 300MB of memory, 2 iterations and 4 threads:
```
kure restore argon2 -m 307200 -i 2 -t 4
```