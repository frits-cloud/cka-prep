# 05 - Cluster upgrade with kubeadm

> Take an etcd snapshot first (exercise 04). Check the installed version with `kubectl get nodes`. You can only upgrade one minor version at a time. See [install-k8s.sh](../install-k8s.sh) for how the packages were installed.

## Task

Upgrade the cluster by one minor version (e.g. 1.31k to 1.32):

1. Find the target version with `kubeadm upgrade plan`.
2. Upgrade `controlplane-1`: kubeadm, `kubeadm upgrade apply`, then kubelet and kubectl.
3. Upgrade `node-1` (drain, upgrade kubeadm, `kubeadm upgrade node`, kubelet, uncordon).
4. Keep a Deployment `upgrade-check` (3 replicas) running in namespace `upgrade` throughout and verify it stays available as nodes are drained.
5. (Optional) Repeat for `node-2` to `node-4`.

## Verify

```bash
k get nodes   # controlplane-1 and node-1 show the new version, all Ready
k get deploy -n upgrade
```

<details>
<summary>Solution</summary>

```bash
k create ns upgrade
k create deploy upgrade-check -n upgrade --image=nginx --replicas=3
```

The apt repository is per minor version, so first change it on each node (example for 1.32):

```bash
sudo sed -i 's|/v1\.[0-9]*/|/v1.32/|' /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
apt-cache madison kubeadm | head
```

Control plane:

```bash
sudo apt-mark unhold kubeadm
sudo apt-get install -y kubeadm=1.32.x-*
sudo apt-mark hold kubeadm
sudo kubeadm upgrade plan
sudo kubeadm upgrade apply v1.32.x

k drain controlplane-1 --ignore-daemonsets
sudo apt-mark unhold kubelet kubectl
sudo apt-get install -y kubelet=1.32.x-* kubectl=1.32.x-*
sudo apt-mark hold kubelet kubectl
sudo systemctl daemon-reload && sudo systemctl restart kubelet
k uncordon controlplane-1
```

Worker (run `k drain` from the control plane, the rest on the node):

```bash
k drain node-1 --ignore-daemonsets --delete-emptydir-data

# on node-1
sudo apt-mark unhold kubeadm kubelet kubectl
sudo apt-get install -y kubeadm=1.32.x-*
sudo kubeadm upgrade node
sudo apt-get install -y kubelet=1.32.x-* kubectl=1.32.x-*
sudo apt-mark hold kubeadm kubelet kubectl
sudo systemctl daemon-reload && sudo systemctl restart kubelet

k uncordon node-1
```

Replace `1.32.x-*` with the real patch version from `apt-cache madison`.

</details>
