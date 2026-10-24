# 10 - Node maintenance and joining a node

Namespace: `maint`

## Task

1. Create a Deployment `app` (image `nginx`, 6 replicas) and a PodDisruptionBudget `app-pdb` with `minAvailable: 4`.
2. Cordon `node-1`, scale `app` to 8 and show that existing pods keep running while new pods avoid `node-1`.
3. Drain `node-1` and watch the evictions respect the PDB. Fix any drain errors (DaemonSets, `emptyDir` volumes, bare pods).
4. Uncordon `node-1`. Observe that pods are not rebalanced automatically.
5. Remove `node-4` from the cluster: drain, `k delete node`, and run `kubeadm reset` on it.
6. Add `node-4` back. Create a new join command on the control plane (the old token expires after 24h) and run it on `node-4`. The format is the same as in [join.pem](../join.pem).
7. Check the certificate expiry on the control plane with `kubeadm certs check-expiration` and renew one certificate.
8. Find the kubelet config, CNI config directory and container runtime socket on a node.

## Verify

```bash
k get nodes
k get pods -n maint -o wide
k get pdb -n maint
```

<details>
<summary>Solution</summary>

```bash
k create ns maint
k create deploy app -n maint --image=nginx --replicas=6
k create pdb app-pdb -n maint --selector=app=app --min-available=4

k cordon node-1
k scale deploy app -n maint --replicas=8      # none of the new pods land on node-1
k get pods -n maint -o wide

k drain node-1 --ignore-daemonsets --delete-emptydir-data
# a bare pod needs --force; the drain waits when the PDB would be violated
k uncordon node-1
```

Remove and re-join:

```bash
k drain node-4 --ignore-daemonsets --delete-emptydir-data
k delete node node-4

# on node-4
sudo kubeadm reset -f
sudo rm -rf /etc/cni/net.d
sudo systemctl restart containerd

# on controlplane-1
kubeadm token create --print-join-command

# on node-4: run the printed command with sudo
sudo kubeadm join <control-plane-ip>:6443 --token <token> \
  --discovery-token-ca-cert-hash sha256:<hash>
```

If you lose the CA hash:

```bash
openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null \
  | openssl dgst -sha256 -hex | sed 's/^.* //'
```

Certificates:

```bash
sudo kubeadm certs check-expiration
sudo kubeadm certs renew apiserver
# restart the control plane static pods afterwards, e.g. by moving the manifests out and back in
```

Node facts:

```bash
sudo cat /var/lib/kubelet/config.yaml
ls /etc/cni/net.d
sudo crictl info | head      # runtime endpoint, usually unix:///run/containerd/containerd.sock
```

</details>
