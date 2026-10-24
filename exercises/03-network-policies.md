# 03 - Network policies

Namespaces: `np-app`, `np-db`, `np-other`

> Network policies are only enforced if the CNI supports them (Calico, Cilium, Weave). Flannel alone does not. Check which CNI your cluster uses first.

## Task

1. Create the three namespaces, label `np-app` with `team=app` and `np-db` with `name=np-db`.
2. In `np-db`, run a pod `db` (image `nginx`, label `app=db`) and expose it with a ClusterIP Service `db` on port 80.
3. In `np-app`, run a pod `client` (image `busybox`, `sleep 3600`, label `app=client`). In `np-other`, run a pod `intruder` the same way.
4. Check that both clients can reach `db.np-db`.
5. Create a default-deny ingress NetworkPolicy in `np-db`.
6. Create a NetworkPolicy `allow-app` in `np-db` that allows ingress to `app=db` pods only from pods in namespaces labelled `team=app` on TCP port 80.
7. Create a NetworkPolicy in `np-app` that denies all egress from `client` except DNS (UDP/TCP 53 to any namespace) and traffic to `np-db` on port 80.

## Verify

```bash
k exec -n np-app client -- wget -qO- -T2 db.np-db        # works
k exec -n np-other intruder -- wget -qO- -T2 db.np-db    # times out
k exec -n np-app client -- wget -qO- -T2 example.com     # fails (egress policy)
```

<details>
<summary>Solution</summary>

```bash
for n in np-app np-db np-other; do k create ns $n; done
k label ns np-app team=app
k label ns np-db name=np-db

k run db -n np-db --image=nginx --labels=app=db
k expose pod db -n np-db --port=80
k run client -n np-app --image=busybox --labels=app=client -- sleep 3600
k run intruder -n np-other --image=busybox -- sleep 3600
```

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-ingress
  namespace: np-db
spec:
  podSelector: {}
  policyTypes: ["Ingress"]
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-app
  namespace: np-db
spec:
  podSelector:
    matchLabels:
      app: db
  policyTypes: ["Ingress"]
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          team: app
    ports:
    - protocol: TCP
      port: 80
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: client-egress
  namespace: np-app
spec:
  podSelector:
    matchLabels:
      app: client
  policyTypes: ["Egress"]
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - { protocol: UDP, port: 53 }
    - { protocol: TCP, port: 53 }
  - to:
    - namespaceSelector:
        matchLabels:
          name: np-db
    ports:
    - { protocol: TCP, port: 80 }
```

Note: in a `from`/`to` list, `namespaceSelector` and `podSelector` in the **same** item are ANDed, in **separate** items they are ORed.

</details>
