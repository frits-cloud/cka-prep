# cka-prep

A small AWS environment, defined with the AWS CDK (Go), for learning and practising for the **CKA (Certified Kubernetes Administrator)** exam: installing a cluster with `kubeadm`, network policies, and the other exam topics.

It is a lab setup. It is not meant for production use.

## What gets deployed

All resources are in one CloudFormation stack (`CdkStack`) in `eu-west-3`:

- A VPC (`10.0.0.0/16`) with 2 public and 2 private subnets across 2 availability zones, an internet gateway and one NAT gateway
- 5 EC2 instances (`t3.medium`, Ubuntu 24.04 LTS) in the private subnets: `controlplane-1` and `node-1` to `node-4`. Each one gets its hostname set at first boot
- A shared security group that allows all traffic between the nodes
- An IAM role that allows access through SSM Session Manager (no SSH, no public IPs) and read/write access to the S3 bucket
- A private S3 bucket that is deleted on `cdk destroy`

The code is in [cdk/](./cdk): `network.go`, `nodes.go` and `s3.go`, wired together in `cdk.go`.

## Prerequisites

### 1. Install Node.js and npm

Install Node.js (which includes npm) from <https://nodejs.org/> or via your package manager, e.g. on Debian/Ubuntu:

```bash
sudo apt install nodejs npm
```

Verify the installation:

```bash
node --version
npm --version
```

### 2. Install the AWS CDK

```bash
sudo npm install -g aws-cdk
```

`sudo` is needed when Node.js was installed via the system package manager (global packages go to a root-owned directory). If you installed Node.js with a version manager such as [nvm](https://github.com/nvm-sh/nvm), omit `sudo`.

Verify the installation:

```bash
cdk --version
```

### 3. Install Go

Install Go from <https://go.dev/doc/install> (the version in [cdk/go.mod](./cdk/go.mod) or newer) and check it:

```bash
go version
```

### 4. Install the AWS CLI and configure a profile

Install the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) and sign in with a profile for the target account, then check it:

```bash
aws sts get-caller-identity --profile <profile>
export AWS_PROFILE=<profile>
```

The account and region are set in `env()` in [cdk/cdk.go](./cdk/cdk.go). Change them to match yours.

To open a shell on a node you also need the [Session Manager plugin](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html).

### 5. Bootstrap the account (once per account and region)

```bash
cd cdk
cdk bootstrap aws://<account-id>/<region>
```

## Usage

```bash
cd cdk
cdk diff      # preview the changes
cdk deploy    # create the environment
```

Connect to a node (find the instance ID in the EC2 console):

```bash
aws ssm start-session --target <instance-id> --region <region> --profile <profile>
```

From there you can practise installing Kubernetes with `kubeadm`, joining the nodes, and working with network policies.

Run the tests with:

```bash
cd cdk
go test ./...
```

## Clean up

The NAT gateway and the instances cost money while they run. Destroy the stack when you are done:

```bash
cd cdk
cdk destroy
```
