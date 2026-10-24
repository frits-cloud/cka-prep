# CKA exercises

Ten hands-on exercises for the lab cluster in this repo (1 control plane, 4 workers, kubeadm). Each file has a task, a way to verify it, and a collapsible solution. Try it without the solution first, and only use `kubernetes.io/docs` as a reference, as in the exam.

| # | Exercise | Exam domain |
|---|----------|-------------|
| 01 | [RBAC](./01-rbac.md) | Cluster Architecture, Installation & Configuration |
| 02 | [Scheduling: taints, affinity, static pods](./02-scheduling.md) | Workloads & Scheduling |
| 03 | [Network policies](./03-network-policies.md) | Services & Networking |
| 04 | [etcd backup and restore](./04-etcd-backup-restore.md) | Cluster Architecture, Installation & Configuration |
| 05 | [Cluster upgrade with kubeadm](./05-cluster-upgrade.md) | Cluster Architecture, Installation & Configuration |
| 06 | [Persistent storage](./06-storage.md) | Storage |
| 07 | [Troubleshoot a broken node](./07-troubleshoot-node.md) | Troubleshooting |
| 08 | [Services and Ingress](./08-services-ingress.md) | Services & Networking |
| 09 | [Deployments, rollouts, HPA, config](./09-workloads.md) | Workloads & Scheduling |
| 10 | [Node maintenance and joining a node](./10-node-maintenance.md) | Cluster Architecture / Troubleshooting |

## Tips

- Set up the usual shortcuts first:
  ```bash
  sudo apt-get install -y bash-completion   # needed for tab completion, not preinstalled on the nodes
  source /usr/share/bash-completion/bash_completion
  alias k=kubectl
  source <(kubectl completion bash) && complete -o default -F __start_kubectl k
  export do="--dry-run=client -o yaml"
  ```
- Use a separate namespace per exercise (given in each task) and delete it afterwards: `k delete ns <name>`.
- Exercises 04, 05, 07 and 10 change the cluster itself. Do them last, and take an etcd snapshot (exercise 04) before 05 and 07.
- Prefer imperative commands (`k create`, `k run`, `k expose`, `k set`) with `$do` to generate YAML quickly.
