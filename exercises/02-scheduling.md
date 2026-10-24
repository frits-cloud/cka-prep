# 02 - Scheduling: taints, affinity, static pods

Namespace: `sched`

## Task

1. Label `node-1` and `node-2` with `disk=ssd` and `node-3` with `disk=hdd`.
2. Taint `node-4` with `dedicated=batch:NoSchedule`.
3. Create a pod `ssd-pod` (image `nginx`) that can only run on nodes labelled `disk=ssd`, using `nodeSelector`.
4. Create a Deployment `web` with 4 replicas (image `nginx`) that uses `requiredDuringSchedulingIgnoredDuringExecution` node affinity to avoid `disk=hdd` nodes, and `preferredDuringSchedulingIgnoredDuringExecution` pod anti-affinity (topology `kubernetes.io/hostname`) to spread the replicas.
5. Create a pod `batch-pod` (image `busybox`, command `sleep 3600`) that tolerates the taint on `node-4` and is placed on `node-4` through a `nodeSelector` on `kubernetes.io/hostname` (not `nodeName`).
6. Create a static pod `static-web` (image `nginx`) on `node-1`.
7. Create a pod `big` that requests `cpu: 100` and observe why it is `Pending`.

## Verify

```bash
k get pods -n sched -o wide
k describe pod big -n sched | grep -A5 Events
```

- `ssd-pod` runs on `node-1` or `node-2`.
- No `web` pod runs on `node-3` or `node-4`.
- `batch-pod` runs on `node-4`.
- A mirror pod `static-web-node-1` exists in the `default` namespace.

<details>
<summary>Solution</summary>

```bash
k create ns sched
k label node node-1 node-2 disk=ssd
k label node node-3 disk=hdd
k taint node node-4 dedicated=batch:NoSchedule
```

`ssd-pod`:

```bash
k run ssd-pod -n sched --image=nginx $do > ssd.yaml
# add under spec:
#   nodeSelector:
#     disk: ssd
k apply -f ssd.yaml
```

`web` Deployment (`k create deploy web -n sched --image=nginx --replicas=4 $do`), add to the pod template spec:

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: disk
          operator: NotIn
          values: ["hdd"]
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        topologyKey: kubernetes.io/hostname
        labelSelector:
          matchLabels:
            app: web
```

`node-4` has no `disk` label, so `NotIn` still matches it, but it is excluded by its taint.

`batch-pod`:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: batch-pod
  namespace: sched
spec:
  nodeSelector:
    kubernetes.io/hostname: node-4
  tolerations:
  - key: dedicated
    operator: Equal
    value: batch
    effect: NoSchedule
  containers:
  - name: c
    image: busybox
    command: ["sleep", "3600"]
```

Static pod: log in to `node-1` and write the manifest to the kubelet's static pod directory (`staticPodPath` in `/var/lib/kubelet/config.yaml`, normally `/etc/kubernetes/manifests`):

```bash
k run static-web --image=nginx $do | sudo tee /etc/kubernetes/manifests/static-web.yaml
```

Unschedulable pod:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: big
  namespace: sched
spec:
  containers:
  - name: c
    image: nginx
    resources:
      requests:
        cpu: "100"
```

```bash
k describe pod big -n sched   # 0/5 nodes are available: Insufficient cpu
```

Clean up: `k taint node node-4 dedicated-`, `k label node node-1 node-2 node-3 disk-`, remove the static pod manifest, `k delete ns sched`.

</details>
