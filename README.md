# Kure

[![PkgGoDev](https://pkg.go.dev/badge/github.com/GGP1/kure)](https://pkg.go.dev/github.com/GGP1/kure)
[![Go Report Card](https://goreportcard.com/badge/github.com/GGP1/kure)](https://goreportcard.com/report/github.com/GGP1/kure)

Kure is a password manager for the command-line that aims to offer a secure (and private) way of operating with sensitive information by reducing the attack surface to its minimum expression.

![Overview](https://user-images.githubusercontent.com/51374959/160211818-b30efbfe-1f1e-44f6-9264-d6faa2f9c0ab.gif)

## Features 

- **Cross-Platform:** Linux, macOS, BSD, Windows and mobile supported.
- **Private:** Completely offline, no connection is established with 3rd parties.
- **Secure:** Each record is encrypted using **AES-GCM 256-bit** and a **unique** password derived using Argon2 (**id** version). The user's master password is **never** stored on disk, it's encrypted and temporarily held **in-memory** inside a protected buffer, which is destroyed immediately after use.
- **Sessions:** Run multiple commands by entering the master password only once. They support setting a timeout and running custom scripts.
- **Portable:** Both Kure and its database compile to binary files and they can be easily carried around in an external device.
- **Easy-to-use:** Intuitive, does not require advanced technical skills.

## Installation

<details>
	<summary>Pre-compiled binaries</summary>
  
Linux, macOS, BSD and Windows pre-compiled binaries can be found [here](https://github.com/GGP1/kure/releases).
</details>

<details>
	<summary>Homebrew (Tap)</summary>

```
brew install GGP1/tap/kure
```
</details>

<details>
	<summary>Scoop (Windows)</summary>

```bash
scoop bucket add GGP1 https://github.com/GGP1/scoop-bucket.git
scoop install GGP1/kure
```
or

```bash
scoop install https://raw.githubusercontent.com/GGP1/scoop-bucket/master/bucket/kure.json
```
</details>

<details>
	<summary>Docker</summary>
	
> For details about persisting the information check the [docker-compose.yml](/docker-compose.yml) file.

```
docker run -it gastonpalomeque/kure sh
```

For a container with limited privileges and kernel capabilities, use:

```
docker run -it --security-opt=no-new-privileges --cap-drop=all gastonpalomeque/kure-secure sh
```
</details>

<details>
	<summary>Mobile phones terminal emulators</summary>

```bash
curl -LO https://github.com/GGP1/kure/releases/download/{version}/{ARM64 file}
tar -xvf {ARM64 file}
mv kure $PATH
```
</details>

<details>
	<summary>Compile from source</summary>

```bash
git clone https://github.com/GGP1/kure
cd kure
make install
```
</details>

## Usage

Further information and examples under [docs/commands](/docs/commands).

<img src="https://user-images.githubusercontent.com/51374959/109055273-b4413180-76bd-11eb-8e71-ae73e7e06522.png" height=550 width=550 />

## Configuration

Out-of-the-box Kure needs no configuration, it creates the configuration file with default values and the database at:

- **Linux, BSD**: `$HOME/.kure`
- **Darwin**: `$HOME/.kure` or `/.kure`
- **Windows**: `%USERPROFILE%/.kure`

However, to store the configuration file elsewhere or use a different one, set the path to it in the `KURE_CONFIG` environment variable.

Head over [configuration](/docs/configuration/configuration.md) for a detailed explanation of the configuration file. Here are some [samples](/docs/configuration/samples/).

### Requirements

- **Linux, BSD**: xsel, xclip, wl-clipboard or Termux:API add-on (termux-clipboard-get/set) to write to the clipboard.
- **macOS**: none.
- **Windows**: none.

## Documentation

This is a simplified version of the documentation, for the full one please visit the [wiki](https://github.com/GGP1/kure/wiki).

### Database

Kure's database is a mantained fork of Bolt ([bbolt](https://github.com/etcd-io/bbolt)), a **key-value** store that uses a single file and a B+Tree structure. The database file is locked when opened, any other simultaneous process attempting to interact with the database will receive a panic.

> The database will always finish all the remaining transactions before closing the connection.

All collections of key/value pairs are stored in **buckets**, five of them are used, one for each type of object and one for storing the authentication parameters. Keys within a bucket must be **unique**, the user will receive an error when trying to create a record with an already used name.

### Data organization

Information isn't really stored inside file folders like we are used to, every record resides at the "root" level but with a path-like key that indicates with which other records it's grouped.

As you may have noticed, the database file isn't encrypted but each one of the records is (and with a unique [password](#master-password)).

> Under the hood, Kure uses *[protocol buffers](https://developers.google.com/protocol-buffers/docs/overview)* (proto 3) for serializing and structuring data.

Names are **case insensitive**, every name's Unicode letter is mapped to its lower case, meaning that "Sample" and "saMple" both will be interpreted as "sample". Spaces within folders and objects names are **allowed**, however, some commands and flags will require the name to be enclosed by double quotes.

### Secret generation

[Atoll](https://www.github.com/GGP1/atoll) library is used for generating cryptographically secure secrets (check the repository documentation for further information).

### Master password

> Remember: the stronger your master password, the harder it will be for the attacker to get access to your information.

Kure uses the [Argon2](https://github.com/P-H-C/phc-winner-argon2/blob/master/argon2-specs.pdf) password hashing function with the **id** version, which utilizes a **32 byte salt** along with the master password and three parameters: *memory*, *iterations* and *threads*. These parameters can modified by the user on registration/restoration. The final key is **256-bit** long.

When encrypting a record, the salt used by Argon2 is randomly generated and appended to the ciphertext, everytime the record is decrypted, the salt is extracted from the end of the ciphertext and used to derive the key. 

Every record is encrypted using a **unique** password, protecting the user against precomputation attacks, such as rainbow tables.

The Argon2id variant with 1 iteration and maximum available memory is recommended as a default setting for all environments. This setting is secure against side-channel attacks and maximizes adversarial costs on dedicated bruteforce hardware.

> If one of the devices that will handle the database has 1GB of memory or less, we recommend setting the *memory* value according to that device's RAM availability.
>
> The command `kure config argon2 test` provides a way of testing the performance implications of different parameter combinations on your device.

### Memory security

Kure encrypts and keeps the master key **in-memory** in a **protected buffer**. When the key is required for an operation, it's **decrypted** and sent into the key derivation function. Right after this, the protected buffer is **destroyed**.

> The library used to perform this operations is called  [memguard](https://github.com/awnumar/memguard). Here are two interesting articles from its author talking about [memory security](https://spacetime.dev/memory-security-go) and [encrypting secrets in memory](https://spacetime.dev/encrypting-secrets-in-memory).

This makes it secure even when the user is into a session and the password resides in the memory.

Finally, it's important to mention that **password comparisons are done in constant time** to avoid side-channel attacks and that **additional sensitive information is wiped after being used** as well. 

### Encryption

Data encryption is done using a **256-bit key**, the symmetric block cipher [AES](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) (Advanced Encryption Standard) along with [GCM](https://en.wikipedia.org/wiki/Galois/Counter_Mode) (Galois/Counter Mode) a cipher mode providing an [authenticated encryption](https://en.wikipedia.org/wiki/Authenticated_encryption) algorithm designed to ensure data authenticity, integrity and confidentiality.

> The national institute of standards and technology (NIST) selected AES as the best algorithm in terms of security, cost, resilience, integrity and surveillance of the algorithm in October 2000.

#### Names aren't encrypted, why?

Although it might be considered a downside and especially if one of the objectives is to make your information as private as possible, there is an explanation.

Encrypting record names would force Kure to use the **exact same key** to do it (it would be virtually impossible to get a match otherwise), making the key susceptible to **precomputation attacks**.

Moreover, the decryption process would be slower, only for the users, preventing them to spend resources on what really matters, the **key derivation function**. 

To sum up, the attacker may (depending on the names) be able to choose which record to attempt a brute-force attack on but using the same key for encryption and "low" argon2 parameters would make it much easier for them to get access to **all** your data.

### Backups

The user can opt to **serve** the database on a **local server** (`kure backup --http --port 8080`) or create a **file** backup (`kure backup --path path/to/file`).

### Restoration

> **Important**: on interrupt signals the database will finish all the remaining transactions before closing the connection.

The database can be restored using the [`kure restore`](https://github.com/GGP1/kure/blob/master/docs/commands/restore.md) command. Every record is decrypted and deleted with the old configuration and re-created with the new one.

The user will be asked to provide a new master password and new argon2 parameters.

### Synchronization

Synchronizing the database between devices can be done in many ways:

> They may introduce new vulnerabilities, use them at your own risk.

1. remotely access a host via ssh with Kure in it.
2. transfer the database file manually.
3. use a file hosting service.

### Sessions

The session command is, essentially, a wrapper of the **root** command and all its subcommands, with the difference that it doesn't exit after executing them. 

This makes sessions great for executing multiple commands passing the master password only **once**, as explained in [master password](#master-password), this is completely secure.

To start a session use [`kure session`](/docs/commands/session.md).

### Two-factor authentication

Kure offers storing two-factor authentication codes in the form of **time-based one-time password (TOTP)**, a variant of the HOTP algorithm that specifies the calculation of a one-time password value based on a representation of the counter as a time factor.

The time-step size used is 30 seconds, a balance between security and usability as specified by [RFC6238](https://tools.ietf.org/html/rfc6238#section-5.2).

> TOTP codes can be either 6, 7 or 8 digits long. The hash algorithm used is SHA1.

Two-factor authentication adds an extra layer of security to your accounts, in case an attacker gets access to the password, he will still need the **constantly refreshing code** to get into the account, making it not impossible but much more complicated.

> Storing an account's password and TOTP key in the same device is not recommended (despite them being encrypted with two totally different keys) as anyone knowing your master password would have access to both.

### Key files

Key files are a [two-factor authentication](#two-factor-authentication) method but for the database, where the user is required not only to provide the correct password but also the path to the key file, which contains a **key** that is **combined with the password** to encrypt the records. 

Using a key file is **optional**, as well as specifying the path to it in the configuration file (if it isn't, it will be requested every time you try to access the database).

> It's safe to store the path to the key file in the configuration file only if it's stored in an external device that must be plugged to log in.

## Caveats and limitations

- Kure cannot provide complete protection against a compromised operating system with malware, keyloggers or viruses.
- There isn't any backdoor or key that can open your database. There is no way of recovering your data if you forget your master password.
- **Windows**: Cygwin/mintty/git-bash aren't supported because they are unable to reach down to the OS API.

## License

Kure is licensed under the Apache-2.0 license. See [LICENSE](/LICENSE).
