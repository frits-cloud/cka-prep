# 01 - RBAC

## Task

1. Create namespaces `rbac-a` and `rbac-b`.
2. In `rbac-a`, create a ServiceAccount `deployer`.
3. Create a Role `deploy-manager` in `rbac-a` that allows `get`, `list`, `create`, `update`, `delete` on `deployments` (apps group) and `get`, `list` on `pods` and `pods/log`.
4. Bind the Role to the ServiceAccount `deployer` with a RoleBinding.
5. Create a ClusterRole `node-reader` that allows `get`, `list`, `watch` on `nodes`, and bind it to the user `jane` with a ClusterRoleBinding.
6. Without using their credentials, check what `deployer` and `jane` can and cannot do.

## Verify

```bash
k auth can-i create deployments -n rbac-a --as system:serviceaccount:rbac-a:deployer   # yes
k auth can-i delete pods -n rbac-a --as system:serviceaccount:rbac-a:deployer          # no
k auth can-i list pods -n rbac-b --as system:serviceaccount:rbac-a:deployer            # no
k auth can-i list nodes --as jane                                                      # yes
k auth can-i delete nodes --as jane                                                    # no
```

## Bonus

Create a client certificate for a user `bob` using the CertificateSigningRequest API, add it to your kubeconfig as a new context, and give `bob` read-only access to all pods in `rbac-b`.

<details>
<summary>Solution</summary>

```bash
k create ns rbac-a
k create ns rbac-b
k create sa deployer -n rbac-a

k create role deploy-manager -n rbac-a \
  --verb=get,list,create,update,delete --resource=deployments.apps
```

An imperative command can't give two resources different verbs, so edit the Role and add the pod rule:

```bash
k edit role deploy-manager -n rbac-a
```

```yaml
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "create", "update", "delete"]
- apiGroups: [""]
  resources: ["pods", "pods/log"]
  verbs: ["get", "list"]
```

```bash
k create rolebinding deploy-manager -n rbac-a --role=deploy-manager --serviceaccount=rbac-a:deployer
k create clusterrole node-reader --verb=get,list,watch --resource=nodes
k create clusterrolebinding node-reader --clusterrole=node-reader --user=jane
```

Bonus:

```bash
openssl genrsa -out bob.key 2048
openssl req -new -key bob.key -subj "/CN=bob" -out bob.csr

cat <<EOF | k apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: bob
spec:
  request: $(base64 -w0 < bob.csr)
  signerName: kubernetes.io/kube-apiserver-client
  usages: ["client auth"]
EOF

k certificate approve bob
k get csr bob -o jsonpath='{.status.certificate}' | base64 -d > bob.crt

k config set-credentials bob --client-certificate=bob.crt --client-key=bob.key --embed-certs
k config set-context bob --cluster=kubernetes --user=bob
k create role pod-reader -n rbac-b --verb=get,list,watch --resource=pods
k create rolebinding bob-pod-reader -n rbac-b --role=pod-reader --user=bob
k auth can-i list pods -n rbac-b --as bob   # yes
```

</details>
