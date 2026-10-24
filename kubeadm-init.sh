sudo kubeadm init \
  --apiserver-advertise-address="$(hostname -I | awk '{print $1}')" \
  --pod-network-cidr=192.168.0.0/16