## Use

`kure config argon2 test [-i iterations] [-m memory] [-t threads]`

## Description

Test how is argon2 going to perform with the parameters passed.

The Argon2id variant with 1 iteration and maximum available memory is recommended as a default setting for all environments. This setting is secure against side-channel attacks and maximizes adversarial costs on dedicated bruteforce hardware.

If one of the devices that will handle the database has lower than 1GB of memory, we recommend setting the memory value to the half of that device RAM availability. Otherwise, default values should be fine.

- Memory: amount of memory allowed for argon2 to use. There is no "insecure" value for this parameter, though clearly the more memory the better. The value is represented in kibibytes, 1 kibibyte = 1024 bytes. Default is 1048576 kibibytes (1024 MB).

- Iterations: number of passes over the memory. The running time depends linearly on this parameter. Again, there is no "insecure value". Default is 1.

- Threads: number of threads number in parallel. Default is the maximum number of logical CPUs usable.

## Flags

|  Name      | Shorthand |     Type      |       Default        |                 Description                   |
|------------|-----------|---------------|----------------------|-----------------------------------------------|
| iterations | i         | uint32        | 1                    | Number of passes over the memory              |
| memory     | m         | uint32        | 1048576              | Amount of memory allowed for argon2 to use    |
| threads    | t         | uint8         | N of CPUs available  | Number of threads running in parallel         |

### Examples

Test using 700MB of memory, 2 iterations and 4 threads:
```
kure config argon2 test -m 716800 -i 2 -t 4
```