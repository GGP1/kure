# Kure

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://godoc.org/github.com/GGP1/kure)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/GGP1/kure)](https://pkg.go.dev/github.com/GGP1/kure)
[![Go Report Card](https://goreportcard.com/badge/github.com/GGP1/kure)](https://goreportcard.com/report/github.com/GGP1/kure)

Kure is a command line password manager.

It also offers storing encrypted files, cards and crypto wallets.

## Table of contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
    * [Commands flags](#commands-flags)
    * [Subcommands](#subcommands)
- [Documentation](#documentation)
    * [How are records stored?](#how-are-records-stored)
    * [Objects](#objects)
    * [Secret generation](#secret-generation)
    * [Master password](#master-password)
    * [Encryption](#encryption)
    * [Backups](#backups)
    * [Sessions](#sessions)
    * [Control over file operations](#control-over-file-operations)
- [Recommendations](#recommendations)
    * [How to choose a secure master password](#how-to-choose-a-secure-master-password)
    * [Two facto authentication](#two-factor-authentication)
- [Dependencies](#dependencies)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Multi-Platform:** Linux, macOS, BSD and Windows supported.
- **Offline:** All the information is handled locally.
- **Password-less:** The user's master password is **never** stored on disk, it's encrypted and stored **in-memory** inside a locked buffer, decrypted when it's required and destroyed right after its use. The key derivation function used is Argon2 with the id version.
- **Secure:**  Data encryption is done by using FIPS-approved cryptographic algorithm AES (Advanced Encryption Standard) along with the GCM (Galois/Counter Mode).
- **Simple and easy to use:** Kure was designed with simplicity in mind.
- **Portable:** Both Kure and the database compile to binary files and they can be easily carried around in an external device.
- **Multiple formats:** entries, bank cards, crypto wallets or files of any type.
- **Customizable:** 

## Installation

No releases yet.

## Configuration

By default Kure will look for the configuration file (.kure.yaml) and create the database (kure.db) in the user home directory ($HOME), unless the path of it is set in an environment variable called `KURE_CONFIG`.

In this file we can also specify where to save our database, modify its name and various default values.

> Paths inside the configuration file must be **absolute**.

Finally, creating a configuration file or reading the one you are using is ultrasimple with the `kure config` command.

*Formats supported*: JSON, TOML, YAML (default), HCL, envfile and Java properties.

[YAML example](/config_example.yaml)

## Usage

For examples and detailed information about each command, please visit [docs/commands](/docs/commands) or execute `kure <command> -h`.

```bash
Usage:
  kure [command]

Available Commands:
  add         Add an entry
  backup      Create database backups
  card        Card operations
  clear       Clear clipboard/terminal or both
  config      Read or create the configuration file
  copy        Copy entry credentials to clipboard
  edit        Edit an entry
  file        File operations
  gen         Generate a random password
  help        Help about any command
  ls          List entries
  rm          Remove an entry
  session     Run a session
  stats       Show database statistics
  wallet      Wallet operations

Flags:
  -h, --help   help for kure

Use "kure [command] --help" for more information about a command.
```

#### Commands flags

Flags might be used with a '=' sign or not.

Here are four ways of running the same command:

`kure add Go -l 15 -f 1,2,3`

`kure add Go -l=15 -f=1,2,3`

`kure add Go --length 15 --format 1,2,3`

`kure add Go --length=15 --format=1,2,3`

```
add <name> [-c custom] [-l length] [-f format] [-i include] [-e exclude] [-r repeat]
backup [http] [port] [path]
card
clear [-b both] [-c clipboard] [-t terminal]
config [-c create] [-p path]
copy <name> [-t timeout] [-u username]
edit <name> [-n name]
gen [-l length] [-f format] [-i include] [-e exclude] [-r repeat] [-q qr]
help
file
ls <name> [-f filter] [-H hide] [-q qr]
rm <name>
session [-p prefix] [-t timeout]
stats
wallet
```

#### Subcommands

[kure add](/docs/commands/add/add.md): phrase.

[kure card](/docs/commands/card/card.md): add, copy, ls, rm.

[kure file](/docs/commands/file/file.md): add, ls, rename, rm, touch.

[kure gen](/docs/commands/gen/gen.md): phrase.

[kure wallet](/docs/commands/wallet/wallet.md): add, copy, ls, rm.

## Documentation

### How are records stored?

Kure's database [Bolt](https://github.com/etcd-io/bbolt) is a **key-value** store that provides an ordered map, which allows easy access and lookup. All collections of key/value pairs are stored in **buckets** within which all keys are stored in byte-sorted order and must be **unique**. Bolt compiles to a single binary.

> A limitation to have into account: Bolt uses a memory-mapped file so the underlying operating system handles the caching of the data. Typically, the OS will cache as much of the file as it can in memory and will release memory as needed to other processes. This means that Bolt can show very high memory usage when working with large databases. However, this is expected and the OS will release memory as needed.

We use four buckets, one for each type of object. There can't be more than one object/folder with the same name/key, the user will be warned if it's trying to create a record with an already used name.

Under the hood, Kure uses [protocol buffers](https://developers.google.com/protocol-buffers/docs/overview) (proto 3) for serializing and structuring data.

For example, adding an entry with the name "Go" will look like:

|                                | kure_entry                 
| -------                        | ------                 
| Key                            | Go
| Value                          | encrypted entry

### Folders

Creating folders couldn't be simpler, all you have to do is include the folder name in the object name when creating it. For example:

`kure add social_media/twitter` will store twitter entry inside the "social_media" folder (spaces within folders and objects names are allowed as well).

#### Objects

```
 ENTRY:                                  CARD:                                  FILE:                                   WALLET:

│   FIELD       │      VALUE       │    │   FIELD       │        VALUE     │    │   FIELD       │        VALUE     │    │   FIELD       │        VALUE     │
│───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│
│ Name          │ x                │    │ Name          │ x                │    │ Name          │ x                │    │ Name          │ x                │
│───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│
│ Username      │ x                │    │ Type          │ x                │    │ Filename      │ x                │    │ Type          │ x                │
│───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│
│ Password      │ x                │    │ Number        │ x                │    │ Size          │ x                │    │ Script Type   │ x                │
│───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│
│ URL           │ x                │    │ CVC           │ x                │    │ Created at    │ x                |    │ Keystore Type │ x                │
│───────────────│──────────────────│    │───────────────│──────────────────│                                            │───────────────│──────────────────│
│ Notes         │ x                │    │ Expires       │ x                │                                            │ Seed Phrase   │ x                │
│───────────────│──────────────────│                                                                                    │───────────────│──────────────────│
│ Expires       │ x                │                                                                                    │ Public Key    │ x                │
                                                                                                                        │───────────────│──────────────────│
                                                                                                                        │ Private Key   │ x                │
```

### Secret generation

For generating secure random secrets we use [Atoll](https://www.github.com/GGP1/atoll) (check repository documentation for further information).

### Master password

Kure won't store your master password **anywhere**, it will be encrypted and handled in-memory in a protected buffer using [memguard](https://github.com/awnumar/memguard). Here are two interesting articles talking about [memory security](https://spacetime.dev/memory-security-go) and [encrypting secrets in memory](https://spacetime.dev/encrypting-secrets-in-memory).

When the key is required for an operation, it's decrypted and sent into a key derivation function called [Argon2](#https://github.com/P-H-C/phc-winner-argon2/blob/master/argon2-specs.pdf) (winner of the Password Hashing Competition in 2015) with the **id** version. Right after this, the protected buffer is destroyed.

> Argon2id key derivation is done with a 32 byte salt along with the master password, the number of logical threads usable and two parameters: *memory* and *iterations* (1024 MB and 1 by default) that can be adapted to your preferences in the configuration file.

The Argon2id variant with 1 iteration and maximum available memory is recommended as a default setting for all environments. This setting is secure against side-channel attacks and maximizes adversarial costs on dedicated bruteforce hardware.

This makes it secure even when the user is into a session and the password resides in the memory.

### Encryption

Encryption is done using [AES](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) (Advanced Encryption Standard), a symmetric block cipher used for information protection, along with [GCM](https://en.wikipedia.org/wiki/Galois/Counter_Mode) (Galois/Counter Mode).

AES became effective as a U.S. federal government standard on May 26, 2002, after approval by the U.S. Secretary of Commerce. It's the first (and only) publicly accessible cipher approved by the U.S. NSA (National Security Agency) for top secret information when used in an [approved cryptographic module](https://apps.nsa.gov/iaarchive/programs/iad-initiatives/cnsa-suite.cfm).

### Backups

The user can opt to serve the database on a local server or create a file backup.

It's really **important** that you keep a backup of your database:

* to avoid losing all your data if you delete the file by accident or if a third person intentionally does it.
* in two or more devices in case one of them stops working.

### Sessions

Sessions are great for executing multiple commands by passing the master password only once, which will be encrypted and stored in a locked buffer until it's needed and destroyed right after it's been used.

To start a session use `kure session`.

You can set a **timeout** using the [-t timeout] flag so it will **automatically close** the session once the time has passed. [Command documentation](/docs/commands/session.md).

### Control over file operations

File operations [add | create | rm] are, by far, the highest memory demanding commands in Kure, that's why we offer the users to regulate their consumption by modifying *buffer* and *semaphore* flags. They could also be used to improve the directive performance.

+ **Buffer**: set the size of the buffer used when reading files. By default it sends the entire file directly into memory.
+ **Semaphore**: maximum number of [goroutines](/docs/commands/file/subcommands/add#goroutines) running concurrently. Default is 1.

In summary, these flags will help you adapt Kure to your requirements.

## Recommendations

#### How to choose a secure master password

Every password manager need at least one password to encrypt/decrypt all the records, this is why it is crucial that you choose a **strong master password** to make it as hard as possible to guess. 

A **good password** is a random combination of upper and lower case letters, numbers and special characters. We recommend choosing a password/passphrase consisting of 20 or more characters (the longer, the better). You should **avoid** picking words that can be found in a dictionary and forget using names or dates of birth.

It's important to note that it shouldn't be stored anywhere and that it must be remembered by the user, forgetting the master password will leave you without access to all your data.

> Forgetful people might find useful writing it down on a paper and saving it somewhere safe.

#### Two-factor authentication

Two-factor authentication is a type, or subset, of multi-factor authentication. It is a method of confirming users' claimed identities by using a combination of **two different factors** (usually 1 and 2): 1. something you know (account credentials), 2. something you have (devices), or 3. something you are.

In case an attacker gets access to the secrets, he will still need the **constantly refreshing code** to get into the account, making it, not impossible, but much more complicated.

## Dependencies

Summarized information about what Kure use them for:

* [github.com/GGP1/atoll](https://www.github.com/GGP1/atoll): generating secrets with a high level of randomness.
* [github.com/atotto/clipboard](https://www.github.com/atotto/clipboard): writing to the clipboard.
* [github.com/awnumar/memguard](https://www.github.com/awnumar/memguard): securely storing sensitive information in memory.
* [github.com/skip2/go-qrcode](https://www.github.com/skip2/go-qrcode): build and display a QR code image on the terminal.
* [github.com/golang/protobuf](https://www.github.com/golang/protobuf) and [google.golang.org/protobuf](https://www.google.golang.org/protobuf): serializing and deserializing objects to store them.
* [github.com/pkg/errors](https//www.github.com/pkg/errors): error handling.
* [github.com/spf13/cobra](https://github.com/spf13/cobra): building cli commands.
* [github.com/spf13/viper](https://github.com/spf13/viper): configuration management.
* [go.etcd.io/bbolt](https://github.com/etcd-io/bbolt): database.
* [golang.org/x/crypto](https://godoc.org/golang.org/x/crypto): encrypting and decrypting records and reading password from the terminal without echoing.

## Contributing

Any contribution is welcome. We appreciate your time and help. Please follow these steps to do it:

> If planning on adding new features, create an issue first.

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