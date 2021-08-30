# armory-cli
[![Go Report Card](https://goreportcard.com/badge/github.com/armory/armory-cli)](https://goreportcard.com/report/github.com/armory/armory-cli) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/gojp/goreportcard/blob/master/LICENSE)

**The CLI for Armory Cloud.**

## Running the CLI
The artifact of building this project is a binary CLI application. To run this project, first build it, then run the executable produced. 
```bash
_$ go build -o armory
_$ ./armory version
INFO[0000] {"version":"development"} 
```

## Libraries
[Cobra](https://github.com/spf13/cobra) - Both a library for creating powerful modern CLI applications and a program to generate applications and command files.

[go-getter](https://github.com/hashicorp/go-getter) - A library for downloading files or directories from various sources using a URL as the primary form of input (local, Git, Mercurial, HTTP, S3, GCP). Not in use yet, but we anticipate leveraging this for features like remote manifest.

