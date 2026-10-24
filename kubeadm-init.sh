sudo kubeadm init \
  --apiserver-advertise-address="$(hostname -I | awk '{print $1}')" \
  --pod-network-cidr=192.168.0.0/16


kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.30.3/manifests/calico.yaml