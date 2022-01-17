# armory-cli
[![Go Report Card](https://goreportcard.com/badge/github.com/armory/armory-cli)](https://goreportcard.com/report/github.com/armory/armory-cli) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/gojp/goreportcard/blob/master/LICENSE)

The CLI for Armory Cloud.

## Installation

The CLI releases can be found on the [releases page](https://github.com/armory/armory-cli/releases/latest)

You can also use the [AVM (armory version manager) utility](https://github.com/armory/avm/releases/latest) to install and update the Armory CLI.

Just download the bin for your os/arch and mark it as executable and place it in your path.

```bash
# Mac OS X Intel
wget https://github.com/armory/avm/releases/download/v1.0.1/avm-darwin-amd64
chmod +x avm-darwin-amd64
mv avm-darwin-amd64 /usr/local/bin/avm
avm install
```

```bash
# Mac OS X M1
wget https://github.com/armory/avm/releases/download/v1.0.1/avm-darwin-arm64
chmod +x avm-darwin-arm64
mv avm-darwin-arm64 /usr/local/bin/avm
avm install
```

### Working Locally With a Mock HTTP Server
Features may not always be developed in the Deploy Engine API. We may also want to test a particular response. This may not be easily 
achieved with unit tests or we may have a need to simulate a state of a deployment during runtime. 

You may run any webserver you like locally which can return a JSON response, but spring-potato is a fine option. You can 
start it with bootRun. The connection to deploy engine requires HTTPS, which means your localhost has to have a valid trusted
SSL Cert. The easiest way to do this is the following:

```
brew install caddy
brew install mkcert
mkcert -install #this makes a trust store on your machine
mkdir ./certs && cd ./certs
mkcert "*.WHATEVER.anythingYouWant" #this is just the CNAME entry you want to use locally
```
Next edit your `/etc/hosts` and add the following:
```aidl
127.0.0.1	specificSubdomain.WHATEVER.anythingYouWant
```
Make a file named: `Caddyfile` in a dir of your choice
```aidl
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
