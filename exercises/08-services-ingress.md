# 08 - Services and Ingress

Namespace: `svc`

## Task

1. Create a Deployment `web` (image `nginx`, 3 replicas) and expose it with a ClusterIP Service `web` on port 80.
2. Create a NodePort Service `web-np` for the same pods with `nodePort: 30080`. Reach it through `<node-ip>:30080`.
3. Create a headless Service `web-headless` and resolve it from a pod. Compare the DNS answer with the ClusterIP Service.
4. Create a second Deployment `api` (image `httpd`) and a Service `api`.
5. Install an ingress controller (for example the [ingress-nginx](https://kubernetes.github.io/ingress-nginx/deploy/#bare-metal-clusters) bare metal manifest, which uses a NodePort Service).
6. Create an Ingress `main` with `ingressClassName: nginx` that routes `/web` to `web:80` and `/api` to `api:80` on host `cka.example.com`, rewriting the path to `/`.
7. Test using the ingress NodePort and a `Host` header.
8. Find the CoreDNS Corefile and the full DNS name of the `web` service.

## Verify

```bash
k get svc,ep,ingress -n svc
curl -H 'Host: cka.example.com' http://<node-ip>:<ingress-nodeport>/web   # nginx welcome page
curl -H 'Host: cka.example.com' http://<node-ip>:<ingress-nodeport>/api   # "It works!"
k run t --image=busybox --rm -it --restart=Never -- nslookup web.svc.svc.cluster.local
```

Node IPs: `k get nodes -o wide`. The nodes have no public IPs (see the [README](../README.md)), so run `curl` from one of the nodes through SSM.

<details>
<summary>Solution</summary>

```bash
k create ns svc
k create deploy web -n svc --image=nginx --replicas=3
k expose deploy web -n svc --port=80
k create deploy api -n svc --image=httpd
k expose deploy api -n svc --port=80

k expose deploy web -n svc --name=web-np --type=NodePort --port=80 $do > np.yaml
# add  nodePort: 30080  under the port, then apply
k apply -f np.yaml

k expose deploy web -n svc --name=web-headless --cluster-ip=None --port=80
k run t --image=busybox --rm -it --restart=Never -- nslookup web-headless.svc   # returns pod IPs
k run t --image=busybox --rm -it --restart=Never -- nslookup web.svc           # returns one ClusterIP
```

Ingress controller:

```bash
k apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.0/deploy/static/provider/baremetal/deploy.yaml
k get svc -n ingress-nginx ingress-nginx-controller
```

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: main
  namespace: svc
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
  - host: cka.example.com
    http:
      paths:
      - path: /web
        pathType: Prefix
        backend:
          service: { name: web, port: { number: 80 } }
      - path: /api
        pathType: Prefix
        backend:
          service: { name: api, port: { number: 80 } }
```

DNS: `k get cm coredns -n kube-system -o yaml`. The name is `web.svc.svc.cluster.local` (`<service>.<namespace>.svc.<cluster-domain>`).

</details>
