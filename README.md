# kure

[![PkgGoDev](https://pkg.go.dev/badge/github.com/GGP1/kure)](https://pkg.go.dev/github.com/GGP1/kure)
[![Go Report Card](https://goreportcard.com/badge/github.com/GGP1/kure)](https://goreportcard.com/report/github.com/GGP1/kure)

Password manager for the command-line that aims to offer a secure and private way of operating with sensitive information by reducing the attack surface to its minimum expression.

![Overview](https://user-images.githubusercontent.com/51374959/160211818-b30efbfe-1f1e-44f6-9264-d6faa2f9c0ab.gif)

## Features

- **Cross-Platform:** Linux, macOS, BSD, Windows and mobile supported.
- **Private:** Self-hosted and completely offline, no connection is established with 3rd parties.
- **Secure:** Each record is encrypted using **AES-GCM** with 256 bit key and a **unique** password derived using Argon2 (**id** version). The user's master password is **never** stored on disk, it's encrypted and temporarily held **in-memory** inside a protected buffer, which is destroyed immediately after use.
- **Sessions:** Run multiple commands by entering the master password only once. They support setting a timeout and running custom scripts.
- **Portable:** Both kure and its database compile to binary files and they can be easily carried around in an external device.
- **Easy-to-use:** Intuitive, does not require advanced technical skills.

## Installation

<details><summary>Pre-compiled binaries</summary>
  
Linux, macOS, BSD, Windows and mobile pre-compiled binaries can be downloaded [here](https://github.com/GGP1/kure/releases).

</details>

<details><summary>Homebrew (Tap)</summary>

```
brew install GGP1/tap/kure
```

</details>

<details><summary>Scoop (Windows)</summary>

```bash
scoop bucket add GGP1 https://github.com/GGP1/scoop-bucket.git
scoop install GGP1/kure
```

or

```bash
scoop install https://raw.githubusercontent.com/GGP1/scoop-bucket/master/bucket/kure.json
```

</details>

<details><summary>Docker</summary>
	
> For details about persisting the information check the [docker-compose.yml](/docker-compose.yml) file.

```
docker run -it gastonpalomeque/kure sh
```

For a container with limited privileges and kernel capabilities, use:

```
docker run -it --security-opt=no-new-privileges --cap-drop=all gastonpalomeque/kure-secure sh
```

</details>

<details><summary>Mobile phones terminal emulators</summary>

```bash
curl -LO https://github.com/GGP1/kure/releases/download/{version}/{ARM64 file}
tar -xvf {ARM64 file}
mv kure $BIN_PATH
```

</details>

<details><summary>Compile from source</summary>

```bash
git clone https://github.com/GGP1/kure
cd kure
make install
```

</details>

## Usage

Further information and examples under [docs/commands](/docs/commands).

<img src="https://github.com/user-attachments/assets/64646f5f-a49d-4dea-97d7-99fab2884158" height=600 width=600 />

## Configuration

Out-of-the-box kure needs no configuration, it creates a file with the default configuration and the database at:

- **Linux, BSD**: `$HOME/.kure`
- **Darwin**: `$HOME/.kure` or `/.kure`
- **Windows**: `%USERPROFILE%/.kure`

However, to store the configuration file elsewhere or use a different one, set the path to it in the `KURE_CONFIG` environment variable.

Head over to the [configuration documentation](/docs/configuration/configuration.md) for a detailed explanation of the configuration file and some [samples](/docs/configuration/samples/).

> [!Note]
> Linux and BSD systems require a utility to write to the clipboard. This could be xsel, xclip, wl-clipboard or the Termux:API add-on.

## Documentation

Learn more about how kure works in the [wiki](https://github.com/GGP1/kure/wiki).

## License

This project is licensed under the Apache-2.0 license. See [LICENSE](/LICENSE).
