# Kure

[![PkgGoDev](https://pkg.go.dev/badge/github.com/GGP1/kure)](https://pkg.go.dev/github.com/GGP1/kure)
[![Go Report Card](https://goreportcard.com/badge/github.com/GGP1/kure)](https://goreportcard.com/report/github.com/GGP1/kure)

Kure is a free and open-source password manager for the command-line.

This project aims to offer the most secure and private way of operating with sensitive data on the terminal, as well as providing a feature-rich and interactive interface to make the user experience simple and enjoyable.

## Features 

- **Cross-Platform:** Linux, macOS, BSD and Windows supported.
- **Offline:** Data is handled locally, no connection is established with 3rd parties.
- **Secure:** Each record is encrypted using **AES-GCM 256-bit** and a **unique** password. Furthermore, the user's master password is **never** stored on disk, it's encrypted and temporarily kept **in-memory** inside a protected buffer, decrypted when it's required and destroyed immediately after it. The key derivation function used is Argon2 with the **id** version.
- **Easy-to-use:** Extremely intuitive and does not require advanced technical skills.
- **Portable:** Both Kure and the database compile to binary files and they can be easily carried around in an external device.
- **Multiple formats:** Store entries, cards and files of any type.

## Table of contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Documentation](#documentation)
  - [Database](#database)
  - [Data organization](#data-organization)
  - [Secret generation](#secret-generation)
  - [Master password](#master-password)
  - [Key file](#key-file)
  - [Memory security](#memory-security)
  - [Encryption](#encryption)
  - [Backups](#backups)
  - [Restoration](#restoration)
  - [Synchronization](#synchronization)
  - [Sessions](#sessions)
  - [Interactive prompt](#interactive-prompt)
  - [Import/Export](#import/export)
  - [Two factor authentication](#two-factor-authentication)
- [Caveats and limitations](#caveats-and-limitations)
- [Contributing](#contributing)
- [License](#license)

## Installation

#### Pre-compiled binaries

Linux, macOS, BSD and Windows pre-compiled binaries can be found [here](https://github.com/GGP1/kure/releases).

#### Homebrew (Tap)
```
brew install GGP1/tap/kure
```

#### Scoop (Windows)
```
scoop bucket add GGP1 https://github.com/GGP1/scoop-bucket.git
scoop install GGP1/kure
```

or

```
scoop install https://raw.githubusercontent.com/GGP1/scoop-bucket/master/bucket/kure.json
```

#### Docker

> The image doesn't have an X server installed and hence, clipboard utilities are not available. Apart from that, everything works as expected.
```
docker run -it gastonpalomeque/kure:0.1.0 sh
```

#### Compìle from source

Requirements: [Go](https://golang.org/doc/install)
```
git clone https://github.com/GGP1/kure
cd kure
make install
```

## Configuration

Out-of-the-box Kure needs no configuration. It sets default values, creates the configuration file and the database at:

- **Linux, BSD**: `$HOME/.kure`
- **Darwin**: `$HOME/.kure` or `/.kure`
- **Windows**: `%USERPROFILE%/.kure`

However, you may want to store the configuration file elsewhere or use a different one, this can be done by setting the path to it in the `KURE_CONFIG` environment variable.

Moving forward to the configuration file itself, in it we can specify the clipboard and session timeouts, the location of the database (if it doesn't exist it will be automatically created), the argon2 parameters, the editor used and whether or not a keyfile is required. Head over [configuration](/docs/configuration/configuration.md) for further details.

> Paths inside the file must be **absolute**.

*Formats supported*: JSON, TOML, YAML, HCL, envfile and Java properties. [Samples](/docs/configuration/samples).

### Requirements

- **Linux, BSD**: xsel, xclip, wl-clipboard or Termux:API add-on (termux-clipboard-get/set) to write to the clipboard.
- **macOS**: none.
- **Windows**: none.

## Usage

Further information and examples about each command under [docs/commands](/docs/commands).

<div>
    <img align="middle" src="https://user-images.githubusercontent.com/51374959/109055273-b4413180-76bd-11eb-8e71-ae73e7e06522.png" height=600 width=600 />
</div>

### Overview

![Overview](https://user-images.githubusercontent.com/51374959/109099553-0efc7c80-7702-11eb-8bab-ad51c004446f.gif)

## Documentation

This is a simplified version of the documentation, for further details, examples and demos please visit the [wiki](https://github.com/GGP1/kure/wiki).

### Database

Kure's database is a mantained fork of Bolt ([bbolt](https://github.com/etcd-io/bbolt)), a **key-value** store that uses a single file and a B+Tree structure. Bolt locks the database file when it's opened, any other simultaneous process attempting to interact with the database will receive a panic.

All collections of key/value pairs are stored in **buckets**, five of them are used, one for each type of object and one for storing the authentication parameters. Keys within a bucket must be **unique**, the user will receive an error when trying to create a record with an already used name.

> The database will always finish all the remaining transactions before closing the connection.

### Data organization

Information isn't really stored inside file folders like we are used to, every record resides at the "root" level but with a path-like key that indicates with which other records it's grouped.

As you may have noticed, the database file isn't encrypted but each one of the records is (and with a unique [password](#master-password)).

> Under the hood, Kure uses *[protocol buffers](https://developers.google.com/protocol-buffers/docs/overview)* (proto 3) for serializing and structuring data.

Names are **case insensitive**, every name's Unicode letter is mapped to its lower case, meaning that "Sample" and "saMple" both will be interpreted as "sample". Spaces within folders and objects names are **allowed**, however, some commands and flags will require the string to be enclosed by double quotes.

### Secret generation

[Atoll](https://www.github.com/GGP1/atoll) library is used for generating cryptographically secure secrets with a high level of randomness (check the repository documentation for further information).

### Master password

> Remember: the stronger your master password, the harder it will be for the attacker to get access to your information.

Kure uses the [Argon2](https://github.com/P-H-C/phc-winner-argon2/blob/master/argon2-specs.pdf) (winner of the Password Hashing Competition in 2015) with the **id** version as the **key derivation function**, it uses a **32 byte salt** along with the master password and three parameters: *memory*, *iterations* and *threads*. The final key is **256-bit** long.

> When encrypting a record, the salt used by Argon2 is randomly generated and appended to the ciphertext, everytime the record is decrypted, the salt is extracted from the end of the ciphertext and used to derive the key. Every record is encrypted using a **unique** password, protecting the user against precomputation attacks, such as rainbow tables.

The Argon2id variant with 1 iteration and maximum available memory is recommended as a default setting for all environments. This setting is secure against side-channel attacks and maximizes adversarial costs on dedicated bruteforce hardware.

If one of the devices that will handle the database has 1GB of memory or less, we recommend setting the *memory* value to the half of that device RAM availability. Otherwise, default values should be fine.

### Key file

Key files are a [two-factor authentication](#two-factor-authentication) method for the database, where the user is required not only to provide the correct password but also the path to the key file, which contains a **key** that is **combined with the password** to encrypt the records. Using a key file is **optional** as well as specifying the path to it in the configuration file (if it isn't specified, it will be requested every time you try to access the database).

> The key file should be as safe as possible and with a limited access.

### Memory security

Kure encrypts and keeps the master key **in-memory** in a **protected buffer**. When the key is required for an operation, it's **decrypted** and sent into the key derivation function. Right after this, the protected buffer is **destroyed**.

> The *"master key"* is the key that is made of the master password, and optionally, the key file.

This makes it secure even when the user is into a session and the password resides in the memory.

It's important to mention that **password comparisons are done in constant time** to avoid side-channel attacks and that **additional sensitive information is wiped after being used** as well. The library used to perform this operations is called  [memguard](https://github.com/awnumar/memguard). Here are two interesting articles from its author talking about [memory security](https://spacetime.dev/memory-security-go) and [encrypting secrets in memory](https://spacetime.dev/encrypting-secrets-in-memory).

#### Useful commands

- [kure config argon2 test](/docs/commands/config/subcommands/argon2/subcommands/test.md) to test argon2 performance.
- [kure config argon2](/docs/commands/config/subcommands/argon2/argon2.md) to see the values under which the password encrypted the records.
- [kure restore argon2](/docs/commands/restore/subcommands/argon2.md) to change the parameters used by argon2.
- [kure restore password](/docs/commands/restore/subcommands/password.md) to change the password.

### Encryption

Data encryption is done using a **256-bit key**, the symmetric block cipher [AES](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) (Advanced Encryption Standard) along with [GCM](https://en.wikipedia.org/wiki/Galois/Counter_Mode) (Galois/Counter Mode) a cipher mode providing an [authenticated encryption](https://en.wikipedia.org/wiki/Authenticated_encryption) algorithm designed to ensure data authenticity, integrity and confidentiality.

> The national institute of standards and technology (NIST) selected AES as the best algorithm in terms of security, cost, resilience, integrity and surveillance of the algorithm in October 2000.

#### Names aren't encrypted, why?

Although it might be considered a downside and especially if one of the objectives is to make your information as private as possible, there is an explanation.

Encrypting record names would force Kure to use the **exact same key** to do it (it would be virtually impossible to get a match otherwise), making the key susceptible to **precomputation attacks**.

Moreover, the decryption process would be slower, only for the users, preventing them to spend resources on what really matters, the **key derivation function**. 

To sum up, the attacker may (depending on the names) be able to choose which record to attempt a brute-force attack on but using the same key for encryption and "low" argon2 parameters would make it much easier for them to get access to **all** your data.

### Backups

We highly encourage to keep a backup of both the database and the key file (in case it's used) in an external hard drive.

The user can opt to **serve** the database on a **local server** (`kure backup --http --port 8080`) or create a **file** backup (`kure backup --path path/to/file`).

### Restoration

> **Important**: on interrupt signals the database will finish all the remaining transactions before closing the connection.

The database can be restored using [`kure restore`](https://github.com/GGP1/kure/blob/master/docs/commands/restore.md). The user will be asked to provide new parameters. Every record is decrypted and deleted with the old configuration and re-created with the new one.

### Synchronization

Synchronizing the database between devices can be done in many ways, they may introduce new vulnerabilities, use them at your own risk:

1. remotely access a host via ssh with Kure in it.
2. transfer the database file manually.
3. use a file hosting service.

### Sessions

The session command is, essentially, a wrapper of the **root** command and all its subcommands, with the difference that it doesn't exit after executing them. This makes sessions great for executing multiple commands passing the master password only **once**, as explained in [master password](#master-password), this is completely secure.

Here's a simplified implementation of [session.go](/cmd/session/session.go):

```go
func runSession(cmd *cobra.Command) error {
          ...

    scanner := bufio.NewScanner(os.Stdin)
    go startSession(cmd, scanner)

    if timeout == 0 {
      block := make(chan struct{})
      <-block
    }

    <-time.After(timeout)
    return nil
}

func startSession(cmd *cobra.Command, scanner *bufio.Scanner) {
  for {
      scanner.Scan()
      args := strings.Split(scanner.Text(), " ")
    
          ...

      r := cmd.Root()
      r.SetArgs(args[:])
      if err := r.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
      }
  }
}
```

To start a session use `kure session`.

You can set a **timeout** using the [-t timeout] flag so it will **automatically close** the session once the time specified has passed. [Command documentation](/docs/commands/session.md).

### Interactive prompt

`kure it` command starts an interactive prompt that behaves depending on the arguments received, it requests the missing information. See its [wiki page](https://github.com/GGP1/kure/wiki/Interactive-prompt) to see more details.

### Import/Export

`kure import` reads other managers' CSV files and stores the entries encrypting them with the master password previously passed.

`kure export` takes Kure's entries and formats them depending on the manager selected to generate a CSV file.

Formats supported: CSV.

Password managers supported:
  - 1Password
  - Bitwarden
  - Keepass/X/XC
  - Lastpass

### Two-factor authentication

Kure offers storing two-factor authentication codes in the form of **time-based one-time password (TOTP)**, a variant of the HOTP algorithm that specifies the calculation of a one-time password value, based on a representation of the counter as a time factor.

The time-step size used is 30 seconds, a balance between security and usability as specified by [RFC6238](https://tools.ietf.org/html/rfc6238#section-5.2).

> TOTP code can be either 6, 7 or 8 digits long. The hash algorithm used is SHA1.

To add a two-factor authentication code it's necessary to have an already stored entry with the same name, this makes 2FA and entries not tightly coupled but slightly related. It's important to mention that this two are encrypted with a different salt.

Two-factor authentication adds an extra layer of security to your accounts. In case an attacker gets access to the secrets, he will still need the **constantly refreshing code** to get into the account making it, not impossible, but much more complicated.

## Caveats and limitations

- Kure cannot provide complete protection against a compromised operating system with malware, keyloggers or viruses.
- There isn't any backdoor or key that can open your database. There is no way of recovering your data if you forget your master password.
- Sharing keys is not implemented as there is no connection with the internet.
- **Windows**: Cygwin/mintty/git-bash aren't supported because they are unable to reach down to the OS API.

## Feedback

We would really appreciate your feedback, feel free to leave your comment [here](https://github.com/GGP1/kure/discussions/categories/feedback).

## Contributing

Any contribution is welcome. We appreciate your time and help. Please follow these steps to do it:

> Do not hesitate to leave any ideas [here](https://github.com/GGP1/kure/discussions/categories/ideas).

1. **Fork** the repository on Github
1. **Clone** your fork - `git clone github.com/<your-username>/kure.git`
1. **Create** your feature branch - `git checkout -b <your-branch>`
1. Make changes, update documentation and tests
1. **Run tests** - `make test-race`
1. **Add** them to staging - `git add .`
1. **Commit** your changes - `git commit -m '<changes>'`
1. **Push** to the branch - `git push origin <your-branch>`
1. Create a **Pull request**

## License

Kure is licensed under the Apache-2.0 license. See [LICENSE](/LICENSE).
