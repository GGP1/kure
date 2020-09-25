# Kure

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://godoc.org/github.com/GGP1/kure)
[![Go Report Card](https://goreportcard.com/badge/github.com/helm/helm)](https://goreportcard.com/report/github.com/GGP1/kure)

Kure is a command line password manager written in pure Go.

This project is not intended for production but for learning purposes.
Although it might be secure and reliable enough, I recommend to use other managers like [1Password][https://1password.com/], [Keypass][https://keepass.info/], [gopass][https://www.gopass.pw/] and others.

## Table of contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
    * [Commands flags](#commands-flags)
- [Documentation](#documentation)
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
- **Secure:** Kure makes use of industry-standard, strong encryption algorithms tested and used by remarkable companies and cryptographers. Also, you can choose whether to store your master password or not, in this last case, you will be asked for it everytime you want to access any information.
- **Dynamic:** Perfect for people that handle passwords frequently.
- **Simple and easy to use:** Kure is simple and intuitive not only to use but to develop and mantain aswell.

## Installation

No releases yet.

## Configuration

Kure by default will create the database and look for the configuration in the user home directory (the same as the GOPATH), unless we set the path to our config file with the global variable `KURE_CONFIG` or by passing the path to the file with the config flag `kure --config path/to/file`.
In this file we can also specify where to save our database, modify its name and set a file with just the master password so kure can read it and use it for encryption and decryption, allowing the user to use external hardware, encrypted disks, etc.

Formats supported: JSON, TOML, YAML, HCL, envfile and Java properties.

[yaml example](/config_example.yaml)

## Usage

For detailed information about each command, please visit [docs/commands](/docs/commands) folder.

```bash
Usage:
  kure [command]

Available Commands:
  add         Adds a new entry to the database
  backup      Create database backups
  card        Add, copy, delete or list cards
  clear       Clear clipboard/terminal
  config      Read or create the configuration file
  copy        Copy password to clipboard
  delete      Delete an entry
  edit        Edit an entry
  gen         Generate a random password
  help        Help about any command
  list        List entries
  stats       Show database statistics
  view        Display all entries on a server
  wallet      Add, copy, delete or list wallets

Flags:
      --config string   config file path
  -h, --help            help for kure

Use "kure [command] --help" for more information about a command.
```

##### Commands flags

```
kure [-h help]
kure add [-c custom | -p phrase] [-s separator] [-l length] [-f format] [-i include]
kure backup [http] [port] [encrypt] [decrypt] [path]
kure card [-a add | -c copy | -d delete | -l list] [-t timeout]
kure clear [-b both] [-c clipboard] [-t terminal]
kure config [-c create] [-p path] 
kure copy <title> [-t timeout]
kure delete <title>
kure edit <title>
kure gen [-l length] [-f format] [-p phrase] [-s separator] [-i include]
kure list <title> [-H hide]
kure stats
kure view [-p port]
kure wallet [-a add | -c copy | -d delete | -l list] [-t timeout]
```

## Documentation

### How are records stored?

Bolt is a key-value store that provides an ordered map, which allows easy access and lookup. All collections of key/value pairs are stored in buckets within which all keys must be unique. The keys are stored in byte-sorted order within a bucket. 
It saves data into a single memory-mapped file on disk. Write-ahead log is not necessary since BoltDB only deals with one file at a time. With copy-on-write, when writing to a page, BoltDB makes updates on the copy of the original page and updates the pointer to point at the new page upon commit.
In case Go version 2 comes with generics, users will be able to create folders and storing entries, cards and wallets in them.

```
bucket         key         value    
 entry
        |- entry title: entry object
 card
        |- card name:   card object
 wallet
        |- wallet name: wallet object
```

##### Objects

```
 ENTRY:                                  CARD:                                   WALLET:

│   FIELD       │      VALUE       │    │   FIELD       │        VALUE     │    │   FIELD       │        VALUE     │
│───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│
│ Title         │ x                │    │ Name          │ x                │    │ Name          │ x                │
│───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│
│ Username      │ x                │    │ Type          │ x                │    │ Type          │ x                │
│───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│
│ Password      │ x                │    │ Number        │ x                │    │ Script Type   │ x                │
│───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│
│ URL           │ x                │    │ CVC           │ x                │    │ Keystore Type │ x                │
│───────────────│──────────────────│    │───────────────│──────────────────│    │───────────────│──────────────────│
│ Notes         │ x                │    │ Expire date   │ x                │    │ Seed Phrase   │ x                │
│───────────────│──────────────────│                                            │───────────────│──────────────────│
│ Expires       │ x                │                                            │ Public Key    │ x                │
                                                                                │───────────────│──────────────────│
                                                                                │ Private Key   │ x                │
```

#### Encryption

Kure hashes user records with SHA-256 and then encrypts it with Bernstein's XChaCha20 symmetric cipher along with Poly1305 message authentication code.
Detailed information [here][https://tools.ietf.org/html/draft-nir-cfrg-chacha20-poly1305-02].

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

Kure is released under the Apache-2.0 license. See [LICENSE](/LICENSE).