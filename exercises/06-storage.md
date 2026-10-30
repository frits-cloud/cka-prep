# 06 - Persistent storage

Namespace: `storage`

## Task

1. Create a StorageClass `manual` with provisioner `kubernetes.io/no-provisioner` and `volumeBindingMode: WaitForFirstConsumer`.
2. Create a v `pv-1` of 1Gi, `ReadWriteOnce`, storage class `manual`, reclaim policy `Retain`, `hostPath` `/mnt/data/pv-1`.
3. Create a PersistentVolumeClaim `data` (500Mi, `ReadWriteOnce`, class `manual`) and note that it stays `Pending` until a pod uses it.
4. Create a pod `writer` (image `busybox`) that mounts `data` at `/data` and writes the hostname to `/data/hello.txt`, then sleeps.
5. Delete the pod, create a pod `reader` that mounts the same claim and prints the file. Confirm the data survived.
6. Delete pod and PVC and look at the PV status. Explain why it is `Released` and not `Available`, and make it `Available` again.
7. Create a pod `shared` with an `emptyDir` volume shared by two containers (one writes, one reads).

## Verify

```bash
k get sc,pv,pvc -n storage
k logs reader -n storage
```

<details>
<summary>Solution</summary>

```bash
k create ns storage
```

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: manual
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-1
spec:
  capacity:
    storage: 1Gi
  accessModes: ["ReadWriteOnce"]
  persistentVolumeReclaimPolicy: Retain
  storageClassName: manual
  hostPath:
    path: /mnt/data/pv-1
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data
  namespace: storage
spec:
  accessModes: ["ReadWriteOnce"]
  storageClassName: manual
  resources:
    requests:
      storage: 500Mi
---
apiVersion: v1
kind: Pod
metadata:
  name: writer
  namespace: storage
spec:
  containers:
  - name: c
    image: busybox
    command: ["sh", "-c", "hostname > /data/hello.txt; sleep 3600"]
    volumeMounts:
    - { name: d, mountPath: /data }
  volumes:
  - name: d
    persistentVolumeClaim:
      claimName: data
```

`reader` is the same with `command: ["cat", "/data/hello.txt"]`. A hostPath volume lives on one node, so if `reader` lands on another node the file is missing. Pin both pods to the same node with `nodeSelector` if needed.

After deleting the PVC, the PV is `Released` because it still has a `claimRef` to the old claim and the `Retain` policy keeps the data. Clear the reference:

```bash
k patch pv pv-1 -p '{"spec":{"claimRef": null}}'
```

Shared emptyDir:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: shared
  namespace: storage
spec:
  containers:
  - name: writer
    image: busybox
    command: ["sh", "-c", "while true; do date >> /shared/log; sleep 5; done"]
    volumeMounts: [{ name: s, mountPath: /shared }]
  - name: reader
    image: busybox
    command: ["sh", "-c", "touch /shared/log; tail -F /shared/log"]
    volumeMounts: [{ name: s, mountPath: /shared }]
  volumes:
  - name: s
    emptyDir: {}
```

</details>
