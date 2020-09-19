# Kure

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://godoc.org/github.com/GGP1/kure)
[![Go Report Card](https://goreportcard.com/badge/github.com/helm/helm)](https://goreportcard.com/report/github.com/GGP1/kure)

Kure is an open-source command line password manager.

This project is not intended for production yet, commands and funcitonalities may change, it's not secure enough and it might crash.

## Table of contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Documentation](#documentation)
    * [Commands](#commands)
    * [How are records stored?](#how-are-records-stored)
    * [Encryption](#encryption)
    * [Randomness](#randomness)
    * [Backup](#backup)
- [Recommendations](#recommendations)
    * [Double-blind passwords](#double-blind-passwords)
- [License](#license)

## Features

- **Multi-Platform support:** Android, DragonFly BSD, FreeBSD, iOS, Linux, macOS, NetBSD, OpenBSD, Plan 9, Solaris, Windows.
- **Local interactivity:** No internet connection required.
- **Secure:** Kure makes use of industry-standard, strong encryption algorithms tested and used by remarkable cryptographers. Also, you can choose whether to store your master password or not, in this last case, you will be asked for it everytime you want to access any information.
- **Dynamic:** Perfect for people that handle passwords frequently.
- **Simple and easy to use:** Kure is designed to be as simple and intuitive as possible not only to use but to develop and mantain aswell.

## Installation

No releases yet.

## Configuration

Kure by default will create the database and will look for the configuration in the user home directory (the same as the GOPATH), unless we set the path to our config file with the global variable `KURE_CONFIG` or by passing the path to the file with the config flag `kure --config path/to/file`.
In this file we can also specify where to save our database, modify its name and set a file with just the master password so kure can read it and use it for encryption and decryption, allowing the user to use external hardware, encrypted disks, etc.

Formats supported: JSON, TOML, YAML, HCL, envfile and Java properties.

*yaml* example:

```
database:
  path: path/to/file
  name: kure

user:
  password: your-password OR password_path: path/to/file

algorithm: aes

entry:
  format: [1,2,3,4,5,6,7,8]

http:
  port: 4000
```

## Documentation

### Commands

For detailed information about each command, please visit [docs/commands](../docs/commands) folder.
```
kure [-h help]
kure add [-c custom | -p phrase] [-s separator] [-l length] [-f format] [-i include]
kure backup [http] [port] [encrypt] [decrypt] [path]
kure card [-a add | -c copy | -d delete | -l list] [-t timeout]
kure clear [-b both] [-c clipboard] [-t terminal]
kure copy <title> [-t timeout]
kure delete <title>
kure edit <title>
kure gen [-l length] [-f format] [-p phrase] [-s separator] [-i include]
kure list <title> [-H hide]
kure view [-p port]
kure wallet [-a add | -c copy | -d delete | -l list] [-t timeout]
```

### How are records stored?

Bolt is a key/value database that uses buckets (collections of key/value pairs) to store records. All keys in a bucket must be unique.

bucket          key         value    
 *entry*
         |- entry title: entry object
 *card*
         |- card name:   card object
 *wallet*
         |- wallet name: wallet object


###### Objects

Entry:                                  Card:                                   Wallet:
```
|   FIELD       |        VALUE     |    |   FIELD       |        VALUE     |    |   FIELD       |        VALUE     |
|---------------|------------------|    |---------------|------------------|    |---------------|------------------|
| Title         | x                |    | Name          | x                |    | Name          | x                |
|---------------|------------------|    |---------------|------------------|    |---------------|------------------|
| Username      | x                |    | Type          | x                |    | Type          | x                |
|---------------|------------------|    |---------------|------------------|    |---------------|------------------|
| Password      | x                |    | Number        | x                |    | Script Type   | x                |
|---------------|------------------|    |---------------|------------------|    |---------------|------------------|
| URL           | x                |    | CVC           | x                |    | Keystore Type | x                |
|---------------|------------------|    |---------------|------------------|    |---------------|------------------|
| Notes         | x                |    | Expire date   | x                |    | Seed Phrase   | x                |
|---------------|------------------|                                            |---------------|------------------|
| Expires       | x                |                                            | Public Key    | x                |
                                                                                |---------------|------------------|
                                                                                | Private Key   | x                |
```

#### Encryption

Kure uses HMAC-SHA256 to hash the user records that then are encrypted by the AES encryption algorithm with the Galois/Counter Mode (GCM) symmetric-key cryptographic mode.
User can modify the algorithm used by specifying which one he prefers in the *config.yaml* file. Right now, AES and twofish are supported only but we expect to expand the list.

#### Randomness

> Note: randomness is a measure of the observer's ignorance, not an inherent quality of a process.

Having this into account, kure uses the crypto/rand package to generate cryptographically secure random numbers and using them to select characters from a pool.
To generate passphrases the procedure is very similar, the same package is used to generate random numbers that determine the length of each word and if the letter is vowel or consonant. We do not use a wordlist as it would give an advantage and make the job easier to the potential attacker.

#### Backups

The user can opt to serve the database file in a local server or doing an encrypted backup of it. More options to be added.

### Recommendations

#### Use double-blind passwords

Money related accounts are, in most cases, the ones with more critical information and we cannot allow any attacker to access them even if they got our master password. To prevent this, we encourage our users to add a non-stored sequence of numbers after the password, for example: longRandomPassword<ID> -> longRandomPassword65874.
What will be stored in the database is longRandomPassword but that isn't the complete one so the attacker won't have access to your accounts even being able to manage all your records. Of course, this sequence of numbers shall be remembered by the user (do not store them anywhere).

## License

Kure is released under the Apache-2.0 license. See [LICENSE](../LICENSE).