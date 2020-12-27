# Kure

[![PkgGoDev](https://pkg.go.dev/badge/github.com/GGP1/kure)](https://pkg.go.dev/github.com/GGP1/kure)
[![Go Report Card](https://goreportcard.com/badge/github.com/GGP1/kure)](https://goreportcard.com/report/github.com/GGP1/kure)

Kure is a simple, secure and reliable command line password manager.

- **Multi-Platform:** Linux, macOS, BSD and Windows supported.
- **Offline:** Data is handled locally, no connection is established with 3rd parties.
- **Secure:** All the records are encrypted using **AES** (Advanced Encryption Standard), a symmetric block cipher along with the **GCM** (Galois/Counter Mode) and a **256-bit** key. Furthermore, the user's master password is **never** stored **anywhere**, it's encrypted and kept **in-memory** inside a locked buffer, decrypted when it's required and destroyed immediately after it. The key derivation function used is Argon2 with the id version.
- **Easy to use:** Kure is extremely intuitive and does not require advanced technical skills.
- **Portable:** Both Kure and the database compile to binary files and they can be easily carried around in an external device.
- **Multiple formats:** Entries, bank cards, files of any type and notes.

## Table of contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Documentation](#documentation)
    * [Database](#database)
    * [Folders](#folders)
    * [Objects](#objects)
    * [Secret generation](#secret-generation)
    * [Master password](#master-password)
    * [Memory security](#memory-security)
    * [Encryption](#encryption)
    * [Backups](#backups)
    * [Restore](#restore)
    * [Synchronization](#synchronization)
    * [Sessions](#sessions)
    * [Import/Export](#import/export)
- [Recommendations](#recommendations)
    * [How to choose a secure master password](#how-to-choose-a-secure-master-password)
    * [Two factor authentication](#two-factor-authentication)
- [Caveats and limitations](#caveats-and-limitations)
- [Dependencies](#dependencies)
- [Contributing](#contributing)
- [License](#license)

## Installation

No releases yet.

## Configuration

Kure will look for the `KURE_CONFIG` environment variable pointing to a configuration file, which at the same time contains the name and path to the database. If the environment variable is not set, the configuration file (.kure.yaml) and the database (kure.db) will be created under the user home directory ($HOME).

Switching between databases is as easy as changing the database name and path in the configuration file. If the file doesn't exist yet, Kure will create it.

> Paths inside the configuration file must be **absolute**.

Finally, creating a configuration file or reading the one you are using is ultrasimple with the `kure config` command.

*Formats supported*: JSON, TOML, YAML (default), HCL, envfile and Java properties.

[YAML example](/config_example.yaml)

## Usage

For examples and detailed information about each command, please visit [docs/commands](/docs/commands) or execute `kure help <command>`/`kure <command> -h`.

For a summarized list of the commands and its flags visit [docs/commands/SUMMARY.md](/docs/commands/SUMMARY.md).

```bash
Usage:
  kure [command]

Available Commands:
  add         Add an entry
  backup      Create database backup
  card        Card operations
  clear       Clear clipboard/terminal (and history) or both
  config      Read or create the configuration file
  copy        Copy entry credentials to the clipboard
  export      Export entries
  file        File operations
  gen         Generate a random password
  help        Help about any command
  import      Import entries
  ls          List entries
  note        Note operations
  restore     Restore the database
  rm          Remove an entry or a directory
  session     Run a session
  stats       Show database statistics

Flags:
  -h, --help   help for kure

Use "kure [command] --help" for more information about a command.
```

## Documentation

### Database

Kure's database is a mantained fork of Bolt ([bbolt](https://github.com/etcd-io/bbolt)), a **key-value** store that provides an ordered map, which allows easy access and lookup. All collections of key/value pairs are stored in **buckets** within which all keys must be **unique**. Bolt compiles to a single binary.

> A limitation to have into account: Bolt uses a memory-mapped file so the underlying operating system handles the caching of the data. Typically, the OS will cache as much of the file as it can in memory and will release memory as needed to other processes. This means that Bolt can show very high memory usage when working with large databases. However, this is expected and the OS will release memory as needed.

We use five buckets, one for each type of object. There can't be more than one object/folder with the same name/key, the user will be warned if it's trying to create a record with an already used name.

Under the hood, Kure uses [protocol buffers](https://developers.google.com/protocol-buffers/docs/overview) (proto 3) for serializing and structuring data. 

For example, adding an *entry* with the name "Go" will look like:

|         | kure_entry                 
| ------- | ------                 
| Key     | Go
| Value   | encrypted entry

### Folders

Creating folders couldn't be simpler, all you have to do is include the name of the folder in the object name when creating it. For example:

`kure add social/twitter` will store "twitter" entry inside the "social" folder. 

Names are **case insensitive**, every name's Unicode letters is mapped to their lower case, meaning that "Sample" and "saMple" both will be interpreted as "sample". Spaces within folders and objects names are **allowed**, however, some commands will require the path to be enclosed by double quotes.

#### Objects

Here is a [list](/docs/objects.md) of Kure's objects.

### Secret generation

For generating secure random secrets we use [Atoll](https://www.github.com/GGP1/atoll) (check repository documentation for further information).

### Master password

> The stronger your master password, the harder it will be for the attacker to get access to your information.

Kure won't store your master password **anywhere**, it will be encrypted and kept in-memory in a protected buffer (see [memory security](#memory-security) for details).

When the key is required for an operation, it's **decrypted** and sent into a key derivation function called [Argon2](https://github.com/P-H-C/phc-winner-argon2/blob/master/argon2-specs.pdf) (winner of the Password Hashing Competition in 2015) with the **id** version. Right after this, the protected buffer is **destroyed**.

This makes it secure even when the user is into a session and the password resides in the memory.

Argon2id key derivation is done with a 32 byte salt along with the master password and three parameters: *memory*, *iterations* (1024 MB and 1 by default) and *threads* (the number of logical threads usable by default), these can be adapted to your preferences using the "kure restore argon2" command. The final key is 256-bit long.

> When encrypting a record, the salt used by Argon2 is randomly generated and appended to the ciphertext, everytime the record is decrypted, the salt is extracted from the end of the ciphertext and used to derive the key.

The Argon2id variant with 1 iteration and maximum available memory is recommended as a default setting for all environments. This setting is secure against side-channel attacks and maximizes adversarial costs on dedicated bruteforce hardware. 

If one of the devices that will handle the database has lower than 1GB of memory, we recommend setting the *memory* value to the half of that device RAM availability. Otherwise, default values should be fine.

##### Useful commands

+ [kure config argon2 test](/docs/commands/config/subcommands/argon2/subcommands/test.md) to **test argon2 performance**. 
+ [kure config argon2](/docs/commands/config/subcommands/argon2/argon2.md) to see the values under which the password encrypted the records.
+ [kure restore argon2](/docs/commands/restore/subcommands/argon2.md) to change the parameters used by argon2.
+ [kure restore password](/docs/commands/restore/subcommands/password.md) to change the password.

### Memory security

> Kure attempts to protect sensitive information stored in-memory but it's not guaranteed.

Go is a managed-memory language, which means we have little control over memory. That's why the approach taken to defend the users from potential attackers is to wipe any bytes that are not longer used and to store sensitive data inside locked buffers using [memguard](https://github.com/awnumar/memguard). Here are two interesting articles from its author talking about [memory security](https://spacetime.dev/memory-security-go) and [encrypting secrets in memory](https://spacetime.dev/encrypting-secrets-in-memory).

> In cases where the data is not immediately needed (like the master password), the data is encrypted before stored in the locked buffer.

### Encryption

Encryption is done using a 256-bit key, the symmetric block cipher [AES](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) (Advanced Encryption Standard) along with [GCM](https://en.wikipedia.org/wiki/Galois/Counter_Mode) (Galois/Counter Mode) a cipher mode providing an [authenticated encryption](https://en.wikipedia.org/wiki/Authenticated_encryption) algorithm designed to ensure both data authenticity (integrity) and confidentiality.

> The national institute of standards and technology (NIST) selected AES as the best algorithm in terms of security, cost, resilience, integrity and surveillance of the algorithm in October 2000.

AES became effective as a U.S. federal government standard, after approval by the U.S. Secretary of Commerce. It's the first (and only) publicly accessible cipher approved by the U.S. NSA (National Security Agency) for top secret information when used in an [approved cryptographic module](https://apps.nsa.gov/iaarchive/programs/iad-initiatives/cnsa-suite.cfm).

#### A depth look into AES

AES 256-bit cipher uses 14 rounds (a substitution and permutation network design with a single collection of steps) of operation for performing encryption and decryption processes. 

AES entire data block is being processed in an identical way during each round. In AES, a plaintext has to travel through *N* number of rounds before producing the cipher. Again, each round comprises four different operations. One operation is permutation and the other three are substitutions. They are SubBytes, ShiftRows, MixColumns, and AddRoundKey.

In AES, all the transformations that are being used in the encryption process will have the inverse transformations that are being used in the decryption process. Each round of the decryption process in AES uses the inverse transformations InvSubBytes(), InvShiftRows() and InvMixColumns().

### Backups

The user can opt to **serve** the database on a **local server** or create a **file** backup with the `kure backup` command.

### Restore

> **WARNING**: interrupting or exiting during a restoring process may cause an irreversible damage to the database data, use it with caution.

Kure provides the capability of restoring the database using different **argon2 parameters** or a **new password**. 

Every record is decrypted and deleted with the old configuration and re-created with the new one.

### Synchronization

Synchronizing the database between devices can be done in two ways:

+ storing the database in a cloud service, having a strong password is enough for it to be safe.
+ transferring the file manually, this last is more tedious but more secure as well.

### Sessions

Sessions are great for executing multiple commands passing the master password only **once**, as explained in [master password](#master-password), this is completely secure.

To start a session use `kure session`.

You can set a **timeout** using the [-t timeout] flag so it will **automatically close** the session once the time has passed. [Command documentation](/docs/commands/session.md).

### Import/Export

`kure import` reads other managers' CSV files and stores the entries encrypting them with the master password previously passed.

`kure export` takes Kure's entries and formats them depending on the manager selected to generate a CSV file.

Formats supported: CSV.

Password managers supported:
  • Bitwarden
  • Keepass
  • Lastpass
  • 1Password

## Recommendations

#### How to choose a secure master password

Every password manager need at least one password to encrypt/decrypt all the records and it is crucial to pick a **strong master password** to make it as hard as possible to guess. 

The password strength depends on the security level that the user requires but in most situations a random combination of upper and lower case letters, numbers and special characters consisting of 20 or more of these (the longer, the better) could be considered strong enough. You should **avoid** picking words that can be found in a dictionary and forget using names or dates of birth.

It's important to note that it shouldn't be stored anywhere and the user must remember it, forgetting the master password will leave you without access to all your data.

#### Two-factor authentication

We higly encourage the usage of Two-factor, a method of confirming users' claimed identities by using a combination of **two different factors** (usually 1 and 2): 1. something you know (account credentials), 2. something you have (devices), or 3. something you are.

It adds another layer of security. In case an attacker gets access to the secrets, he will still need the **constantly refreshing code** to get into the account making it, not impossible, but much more complicated.

## Caveats and limitations

+ Kure cannot provide complete protection against a compromised operating system with malware, keyloggers or viruses.
+ There isn't any backdoor or key that can open your database. There is no way of recovering your data if you forget your master password.
+ **Windows users** have to clean the terminal history manually, however, closing and opening a new one is a quick solution as it stores the session history only. Moreover, Cygwin/mintty/git-bash are not supported on this platform because they are unable to reach down to the OS API.
+ Sharing keys is not implemented as there is no connection with the internet.
+ Kure doesn't validate passwords, is up to the user to use strong ones.

## Dependencies

|                            Dependency                                 |                  License                |                             Used for                                |
|-----------------------------------------------------------------------|-----------------------------------------|---------------------------------------------------------------------|
| [github.com/GGP1/atoll](https://www.github.com/GGP1/atoll)            | MIT License                             | Generating secrets with a high level of randomness                  |
| [github.com/atotto/clipboard](https://www.github.com/atotto/clipboard)| BSD-3-Clause License                    | Writing to the clipboard                                            |
| [github.com/awnumar/memguard](https://www.github.com/awnumar/memguard)| Apache-2.0 License                      | Store sensititive information in-memory securely                    |
| [github.com/skip2/go-qrcode](https://www.github.com/skip2/go-qrcode)  | MIT License                             | Build and display a QR code image on the terminal                   |
| [github.com/golang/protobuf](https://www.github.com/golang/protobuf)  | BSD-3-Clause License                    | Serializing and deserializing objects to store them                 |
| [github.com/pkg/errors](https://github.com/pkg/errors)                | BSD-2-Clause License                    | Error handling                                                      |
| [github.com/spf13/cobra](https://github.com/spf13/cobra)              | Apache-2.0 License                      | Building CLI commands                                               |
| [github.com/spf13/viper](https://github.com/spf13/viper)              | MIT License                             | Configuration management                                            |
| [go.etcd.io/bbolt](https://github.com/etcd-io/bbolt)                  | MIT License                             | Database                                                            |
| [golang.org/x/crypto](https://godoc.org/golang.org/x/crypto)          | BSD 3-Clause "New" or "Revised" License | Encryption and reading passwords from the terminal without echoing  |

## Feedback

We would really appreciate your feedback, feel free to leave your comment [here](https://github.com/GGP1/kure/discussions/categories/feedback).

## Contributing

Any contribution is welcome. We appreciate your time and help. Please follow these steps to do it:

> Do not hesitate to leave any ideas [here](https://github.com/GGP1/kure/discussions/categories/ideas).

1. **Fork** the repository on Github
2. **Clone** your fork by executing: `git clone github.com/<your_username>/kure.git`
3. **Create** your feature branch (`git checkout -b <your-branch>`)
4. Make changes and **run tests** (`go test ./... -race -p 1`)
5. **Add** them to staging (`git add .`)
6. **Commit** your changes (`git commit -m '<changes>'`)
7. **Push** to the branch (`git push origin <your-branch>`)
8. Create a **Pull request**

## License

Kure is licensed under the Apache-2.0 license. See [LICENSE](/LICENSE).