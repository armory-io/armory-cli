# armory-cli
[![Go Report Card](https://goreportcard.com/badge/github.com/armory/armory-cli)](https://goreportcard.com/report/github.com/armory/armory-cli) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/gojp/goreportcard/blob/master/LICENSE)

The CLI for Armory CD-as-a-Service.

# Installation

Install `armory` on Mac OS using [Homebrew](https://brew.sh/):

```shell
brew tap armory-io/armory
brew install armory-cli
```

To install `armory` on Windows or Linux, run the following:

```shell
curl -sL go.armory.io/get-cli | bash
```

The script will install `armory` and `avm`. You can use `avm` (**A**rmory **V**ersion **M**anager) to manage your `armory` version. 

The CLI releases can be found on the [releases page](https://github.com/armory/armory-cli/releases/latest)

# Release

To make a public release, create a semantic version tag and push it to this repository:
```shell
git tag v1.10.0
git push --tags
```

# Development

## Initial setup

After cloning repository, make sure scripts module is initialized:

```shell
 git submodule update --init --recursive
```

## Changing Environments

During your development cycles, you'll likely want to try out the commands you're building. Aside from executing your unit tests,
you can run the application, either from the command line or from your IDE (i.e. IntelliJ IDEA).

The CLI respects options in your command line environment as well as receiving them as options when executing a particular command.
As an example you can set the address of the environment to work with in two ways:

```
ARMORY_ADDR=https://api.dev.cloud.armory.io && armory login
//OR
armory --addr https://api.dev.cloud.armory.io login
```
You may find it useful to add aliases to your bash profile:
```
alias ac_dev='export ARMORY_ADDR=https://api.dev.cloud.armory.io'
alias ac_staging='export ARMORY_ADDR=https://api.staging.cloud.armory.io'
alias ac_prod='export ARMORY_ADDR=https://api.cloud.armory.io'
```
You can switch environments easily with the above.

## Building and Running Locally - Armory Cloud Dev
### IntelliJ IDEA
Edit Configurations > Add Configuration (+) of type `Go Build`
```
Run Kind: Package
Package path: github.com/armory/armory-cli
Output dir: (unset) [check] Run after build 
Working directory: /Users/..../armory-io/armory-cli
Program arguments: `quick-start --addr https://api.dev.cloud.armory.io -a "JWT TOKEN HERE"
```
You can substitute any command for `quick-start`. Using the `-a` option isn't necessary if you `armory login` first to set a valid token 
in your `~.armory` folder.

The `--addr` option is important. This tells the CLI which environment you're attempting to use, which will override the default, production, environment.

### Run in your console
```
go build -o armory && chmod +x armory && ./armory --addr https://api.dev.cloud.armory.io login
```
Here we're building, changing the permission of our build to make it executable and then logging in. On subsequent
commands, you can use `./armory quick-start` without the need to build again unless you've made changes.

### Working Locally With a Mock HTTP Server
Features may not always be developed in Deploy Engine API. We may also want to test a particular response, which may not be easily 
achieved with unit tests. Or we may want to simulate a state of a deployment during runtime. 

You may run any webserver you like locally which can return a JSON response, but spring-potato is a fine option. You can 
start it with bootRun. The connection to deploy engine requires HTTPS, which means your localhost has to have a valid trusted
SSL Cert. The easiest way to do this is the following:

```shell
brew install caddy
brew install mkcert
mkcert -install #this makes a trust store on your machine
mkdir ./certs && cd ./certs
mkcert "*.WHATEVER.anythingYouWant" #this is just the CNAME entry you want to use locally
```
Next edit your `/etc/hosts` and add the following:
```hosts
127.0.0.1	specificSubdomain.WHATEVER.anythingYouWant
```
Make a file named: `Caddyfile` in a dir of your choice
```
specificSubdomain.WHATEVER.anythingYouWant {
  tls ./_wildcard.WHATEVER.anythingYouWant.pem ./_wildcard.WHATEVER.anythingYouWant-key.pem
  reverse_proxy localhost:8080 {
  header_up Host                {host}
      header_up Origin              {host}
      header_up X-Real-IP           {remote}
      header_up X-Forwarded-Host    {host}
      header_up X-Forwarded-Server  {host}
      header_up X-Forwarded-Port    {port}
      header_up X-Forwarded-For     {remote}
      header_up X-Forwarded-Proto   {scheme}
  }
}

```
While in that directory, execute `caddy run` and it will automatically pick up your configuration, otherwise if you're not in
the directory use the `--config <locationOfYourConfig>`. You should not see any errors. Make sure your spring potato app
is running at `localhost:8080`  

## Building Docker images
You can build a preview Docker image with `make release`. It can be optionally published to our private Artifactory Docker registry:
```shell
make PUSH=true release
```
