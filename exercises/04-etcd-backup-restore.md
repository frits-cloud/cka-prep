# 04 - etcd backup and restore

Run on `controlplane-1`.

## Task

1. Find the etcd endpoint, certificates and data directory (hint: the etcd static pod manifest).
2. Create a namespace `etcd-test` with a Deployment `before-backup` (image `nginx`).
3. Take an etcd snapshot to `/opt/backup/etcd-snapshot.db` and verify it with `etcdutl snapshot status`.
4. Create another Deployment `after-backup` in `etcd-test`.
5. Restore the snapshot into a new data directory `/var/lib/etcd-restore` and make etcd use it.
6. Confirm that `before-backup` exists and `after-backup` is gone.

## Verify

```bash
k get deploy -n etcd-test
sudo ETCDCTL_API=3 etcdctl --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key endpoint health
```

<details>
<summary>Solution</summary>

If `etcdctl`/`etcdutl` are not installed, download them from the [etcd releases](https://github.com/etcd-io/etcd/releases) (match the etcd version in the manifest) or use `sudo apt-get install -y etcd-client`.

```bash
sudo grep -E 'listen-client|--cert-file|--key-file|--trusted-ca-file|--data-dir' /etc/kubernetes/manifests/etcd.yaml

k create ns etcd-test
k create deploy before-backup -n etcd-test --image=nginx

sudo mkdir -p /opt/backup
sudo ETCDCTL_API=3 etcdctl --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key \
  snapshot save /opt/backup/etcd-snapshot.db

sudo etcdutl snapshot status /opt/backup/etcd-snapshot.db -w table

k create deploy after-backup -n etcd-test --image=nginx
```

Restore (`etcdutl` on etcd 3.6 and newer, `etcdctl snapshot restore` on older versions; both accept the same flags here):

```bash
sudo etcdutl snapshot restore /opt/backup/etcd-snapshot.db --data-dir=/var/lib/etcd-restore
```

Point etcd at the new directory by editing the hostPath of the `etcd-data` volume in `/etc/kubernetes/manifests/etcd.yaml`:

```yaml
  - hostPath:
      path: /var/lib/etcd-restore
      type: DirectoryOrCreate
    name: etcd-data
```

The kubelet recreates the etcd pod. This takes a minute, and `kubectl` will fail meanwhile. Watch with `sudo crictl ps | grep etcd`.

```bash
k get deploy -n etcd-test   # only before-backup
```

</details>
