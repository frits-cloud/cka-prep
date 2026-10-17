# cka-prep
CKA (Certified Kubernetes Administrator) prep environment

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
