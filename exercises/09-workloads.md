# 09 - Deployments, rollouts, HPA and configuration

Namespace: `wl`

## Task

1. Create a Deployment `app` (image `nginx:1.25`, 3 replicas) with `maxSurge: 1` and `maxUnavailable: 0`.
2. Update the image to `nginx:1.26`, watch the rollout, then roll back to the previous revision. Record the change cause in the rollout history.
3. Do a rollout to a non-existing image tag, see it get stuck, and undo it.
4. Create a ConfigMap `app-config` with `COLOR=blue` and `MODE=prod`, and a Secret `app-secret` with `PASSWORD=s3cr3t`.
5. Inject `COLOR` as an env var, all of `app-secret` as env vars, and mount `app-config` as files in `/etc/config`.
6. Add resource requests (`cpu: 100m`, `memory: 64Mi`), limits (`cpu: 200m`, `memory: 128Mi`), a readiness probe (HTTP GET `/` on 80) and a liveness probe.
7. Install metrics-server (it needs `--kubelet-insecure-tls` in this lab) and create an HPA for `app`: min 2, max 6, target 50% CPU.
8. Generate load and watch the HPA scale up.
9. Create a CronJob `hello` that runs every minute with `busybox` and `date`, keeping the last 2 successful jobs. Create a Job `once` with `completions: 3` and `parallelism: 2`.

## Verify

```bash
k rollout history deploy/app -n wl
k exec -n wl deploy/app -- env | grep -E 'COLOR|PASSWORD'
k exec -n wl deploy/app -- ls /etc/config
k top pods -n wl
k get hpa -n wl -w
k get cronjob,job -n wl
```

<details>
<summary>Solution</summary>

```bash
k create ns wl
k create deploy app -n wl --image=nginx:1.25 --replicas=3
k patch deploy app -n wl -p '{"spec":{"strategy":{"type":"RollingUpdate","rollingUpdate":{"maxSurge":1,"maxUnavailable":0}}}}'

k set image deploy/app -n wl nginx=nginx:1.26
k annotate deploy/app -n wl kubernetes.io/change-cause="upgrade to 1.26"
k rollout status deploy/app -n wl
k rollout history deploy/app -n wl
k rollout undo deploy/app -n wl

k set image deploy/app -n wl nginx=nginx:does-not-exist
k rollout status deploy/app -n wl --timeout=30s    # stuck
k rollout undo deploy/app -n wl
```

Config:

```bash
k create cm app-config -n wl --from-literal=COLOR=blue --from-literal=MODE=prod
k create secret generic app-secret -n wl --from-literal=PASSWORD=s3cr3t
k edit deploy app -n wl
```

```yaml
      containers:
      - name: nginx
        image: nginx:1.25
        env:
        - name: COLOR
          valueFrom:
            configMapKeyRef: { name: app-config, key: COLOR }
        envFrom:
        - secretRef: { name: app-secret }
        resources:
          requests: { cpu: 100m, memory: 64Mi }
          limits: { cpu: 200m, memory: 128Mi }
        readinessProbe:
          httpGet: { path: /, port: 80 }
        livenessProbe:
          httpGet: { path: /, port: 80 }
          initialDelaySeconds: 5
        volumeMounts:
        - { name: config, mountPath: /etc/config }
      volumes:
      - name: config
        configMap: { name: app-config }
```

metrics-server and HPA:

```bash
k apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
k patch deploy metrics-server -n kube-system --type=json \
  -p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'

k expose deploy app -n wl --port=80
k autoscale deploy app -n wl --min=2 --max=6 --cpu-percent=50
k run load -n wl --image=busybox --restart=Never -- sh -c 'while true; do wget -qO- http://app >/dev/null; done'
```

Jobs:

```bash
k create cronjob hello -n wl --image=busybox --schedule='* * * * *' -- date
k patch cronjob hello -n wl -p '{"spec":{"successfulJobsHistoryLimit":2}}'
k create job once -n wl --image=busybox $do -- echo done > job.yaml
# add under spec: completions: 3 and parallelism: 2
k apply -f job.yaml
```

</details>
