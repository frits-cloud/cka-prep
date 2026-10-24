#!/usr/bin/env bash
# Open an SSM Session Manager shell on a cka-prep node by name, e.g. ./ssm-shell.sh node-1
set -euo pipefail

name="${1:?usage: $0 <node-name>}"
region="${AWS_REGION:-eu-west-3}"

# CDK sets the Name tag to the construct path, e.g. CdkStack/node-1
id="$(aws ec2 describe-instances \
  --region "$region" \
  --filters "Name=tag:Project,Values=cka-prep" \
            "Name=tag:Name,Values=*/${name}" \
            "Name=instance-state-name,Values=running" \
  --query 'Reservations[].Instances[].InstanceId' \
  --output text)"

if [[ -z "$id" || "$id" == *$'\t'* ]]; then
  echo "no single running instance found for '$name' in $region" >&2
  exit 1
fi

echo "connecting to $name ($id)"
exec aws ssm start-session --region "$region" --target "$id"
