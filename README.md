# Kuhnuri AWS CLI

Kuhnuri command line tool to run conversions in AWS.

## Installing

Requirements:

* [Go 1.12 or newer](https://golang.org/dl/)

```bash
$ go get -u github.com/kuhnuri/kuhnuri-cli/kuhnuri
```

## Usage

Requires a running [Kuhnuri environment](https://github.com/kuhnuri/kuhnuri-cdk).

```bash
$ kuhnuri -i INPUT -f TRANSTYPE [-o OUTPUT] [--api API]
```
or
```bash
$ kuhnuri -p PROJECT [-o OUTPUT] [--api API]
``` 

<dl>
  <dt><code>INPUT</code></dt>
  <dd>Input URL or local file</dd>
  <dt><code>TRANSTYPE</code></dt>
  <dd>Transtype name</dd>
  <dt><code>PROJECT</code></dt>
  <dd>DITA-OT project file</dd>
  <dt><code>OUTPUT</code></dt>
  <dd>Output URL or local directory. Optional</dd>
  <dt><code>API</code></dt>
  <dd>Kuhnuri API URL. Optional</dd>
</dl>

If `INPUT` is a local file, contents of the parent directory is zipped and uploaded to a temporary AWS S3 resource.
 
If `OUTPUT` is a local directory, the output ZIP will be uploaded and extracted there.

## Configuration

Default options can be configured in `.kuhnurirc` file

```properties
api = https://example.com/
```

The configuration file is searched for in order:

* current directory or any ancestor directory
* user home directory

The first `.kuhnurirc` file found is used.

## Donating

Support this project and others by [@jelovirt](https://github.com/jelovirt) via [Paypal](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=jarno%40elovirta%2ecom&lc=FI&item_name=Support%20Open%20Source%20work&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donate_LG%2egif%3aNonHosted).

## License

Kuhnuri AWS CLI is licensed for use under the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0).
