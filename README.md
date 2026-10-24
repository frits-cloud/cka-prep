# cka-prep

A small AWS lab, defined with the AWS CDK (Go), for practising for the **CKA (Certified Kubernetes Administrator)** exam: installing a cluster with `kubeadm`, network policies and the other exam topics. It is not meant for production use.

## What gets deployed

One CloudFormation stack (`CdkStack`) in `eu-west-3`:

- A VPC with 2 public and 2 private subnets across 2 AZs and one NAT gateway
- 5 EC2 instances (`t3.medium`, Ubuntu 24.04) in the private subnets: `controlplane-1` and `node-1` to `node-4`
- A security group that allows all traffic between the nodes
- An IAM role for SSM Session Manager (no SSH, no public IPs) with read/write access to a private S3 bucket that is deleted on `cdk destroy`

The code is in [cdk/](./cdk), wired together in [cdk.go](./cdk/cdk.go).

## Prerequisites

- [Node.js](https://nodejs.org/) and the CDK: `sudo npm install -g aws-cdk` (omit `sudo` if you use nvm)
- [Go](https://go.dev/doc/install), version as in [cdk/go.mod](./cdk/go.mod) or newer
- The [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) with a profile for the target account (`export AWS_PROFILE=<profile>`), and the [Session Manager plugin](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html) to open shells
- An EC2 key pair named `cka-prep` in the region, which the worker nodes reference

Set your account and region in `env()` in [cdk/cdk.go](./cdk/cdk.go), then bootstrap once per account and region:

```bash
cd cdk
cdk bootstrap aws://<account-id>/<region>
```

## Usage

```bash
cd cdk
cdk diff      # preview the changes
cdk deploy    # create the environment
go test ./... # run the tests
```

Open a shell on a node by name:

```bash
./scripts/ssm-shell.sh node-1
```

In VS Code, run the task **Tasks: Run Task → shell: all nodes** ([.vscode/tasks.json](./.vscode/tasks.json)) to open a session on every node at once. They show up as split panes in one terminal group. You can also run a single node, e.g. `shell: controlplane-1`. The Session Manager plugin and valid AWS credentials (`AWS_PROFILE`) are needed in the environment VS Code runs in.

### Run a script on all nodes

[remote/](./remote) finds every running instance tagged `Project=cka-prep` and runs a script on all of them in parallel with SSM Run Command. It needs no SSH or key, and the script runs as root. By default it runs [install-k8s.sh](./install-k8s.sh), which installs containerd, kubeadm, kubelet and kubectl.

```bash
cd remote
go run .                           # all nodes
go run . -nodes controlplane-1     # selected nodes
```

Other flags: `-region`, `-script` and `-timeout`. Output is printed per node when it finishes (SSM keeps only the last 24,000 characters). After that, create the cluster with `kubeadm init` on `controlplane-1` and `kubeadm join` on the workers.

## Clean up

The NAT gateway and instances cost money while they run:

```bash
cd cdk
cdk destroy
```
