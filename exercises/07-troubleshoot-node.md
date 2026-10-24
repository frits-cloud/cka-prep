# 07 - Troubleshoot a broken node and control plane

This exercise needs a "saboteur" (you in a second terminal, or a friend). Break things, then fix them using only logs and the docs.

## Task

Break each of the scenarios below one at a time, diagnose from the symptoms and fix it.

1. **Stopped kubelet**: on `node-2` run `sudo systemctl stop kubelet`. The node becomes `NotReady`. Fix it.
2. **Wrong kubelet config**: on `node-3` change the `clientCAFile` in `/var/lib/kubelet/config.yaml` to a wrong path and restart kubelet. Find the error in `journalctl` and fix it.
3. **Broken scheduler**: on `controlplane-1` change a flag in `/etc/kubernetes/manifests/kube-scheduler.yaml` (for example `--kubeconfig=/etc/kubernetes/scheduler-wrong.conf`). New pods stay `Pending`. Find why with `crictl` and the logs in `/var/log/pods`, then fix it.
4. **Crashing pods**: create a Deployment with a typo in the image name (`ImagePullBackOff`) and another whose command exits with 1 (`CrashLoopBackOff`). Fix both.
5. **Service without endpoints**: create a Deployment `app` (label `app=app`) and a Service whose selector says `app=wrong`. Find out why there are no endpoints.
6. **DNS**: scale CoreDNS to 0, observe that `nslookup kubernetes` fails in a pod, scale it back.

## Verify

```bash
k get nodes
k get pods -A
k get endpoints
k run dns-test --image=busybox --rm -it --restart=Never -- nslookup kubernetes.default
```

<details>
<summary>Toolbox and solutions</summary>

Where to look:

```bash
k describe node <node>                        # Conditions and Events
k get events -A --sort-by=.lastTimestamp
k logs <pod> [-p] [-c container]              # -p = previous crash
sudo systemctl status kubelet
sudo journalctl -u kubelet --no-pager | tail -50
sudo crictl ps -a
sudo crictl logs <container-id>
ls /var/log/pods /var/log/containers
```

1. `sudo systemctl enable --now kubelet`
2. `journalctl -u kubelet` shows the bad file or path. Fix `/var/lib/kubelet/config.yaml` (or the drop-in in `/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf`), then `sudo systemctl daemon-reload && sudo systemctl restart kubelet`.
3. Static pod manifests are watched by the kubelet, so saving a fixed manifest is enough. `sudo crictl ps -a | grep scheduler` and `crictl logs` show the error.
4. `k describe pod` shows the pull error. `k set image deploy/<name> <container>=nginx`. For the crash, `k logs` and `k edit deploy`.
5. `k get endpoints <svc>` is empty. Compare `k get svc <svc> -o yaml` selector with `k get pods --show-labels`, then `k edit svc <svc>`.
6. `k scale deploy coredns -n kube-system --replicas=2`. Also check `k get svc -n kube-system kube-dns` and `/etc/resolv.conf` in the pod.

</details>
