# VCLI

Some handy CLI utilities for local development

## Usage
### Prerequisites
- Go installed
- Go environment properly configured (GOPATH,GOROOT)
- Create `vcli.yaml` config file in your home directory
### Install
```shell
go get github.com/vladcar/vcli
```
### Features
```shell
user@user ~ % vcli
Vlad's personal CLI utils

Usage:
  vcli [command]

Available Commands:
  awsconf     Assume IAM role using SSO credentials and export temporary credentials to your shell dotfile e.g. .zshenv . Prior SSO login required
  help        Help about any command
  server      Start basic http server

Flags:
      --config string   config file (default is $HOME/vcli.yaml) (default "/Users/vladc/vcli.yaml")
  -h, --help            help for vcli

Use "vcli [command] --help" for more information about a command.
```

#### Example `vcli.yaml`
```yaml
aws:
  some-dev:
    awsProfile: some-dev
    roleArn: <your AWS SSO role ARN>
  some-prod:
    awsProfile: some-prod
    roleArn: <your AWS SSO role ARN>
  personal:
    awsProfile: personal
    roleArn: <your role ARN>
```

#### Example `awsconf` usage
```shell
aws sso login --profile my-profile
...<complete sso login>...

vcli awsconf --profile my-profile --region eu-central-1
```
Output should be like this
```shell
Profile: my-profile
Role: arn:aws:iam::<ACCOUNT_ID>:role/aws-reserved/sso.amazonaws.com/<REGION>/<ROLE NAME>
Region: <REGION>
Shell dotfile: .zshenv
role assumed, session id: FAGsP1MhfpBE2M3LqMHaX
exporting AWS credentials to:  /Users/user/.zshenv
done
To get started with using AWS you may need to restart your current shell.
This would reload your environment to include latest exported AWS credentials.

To configure your current shell, run:
source $HOME/.zshenv
```
