#!/usr/bin/env bash

green="\e[0;92m"
red="\e[0;91m"
reset="\e[0m"
if [ "$EUID" != 0 ]; then
    echo -e "${red}Must be run as root"
    exit 1
fi

echo -e "This script will install Kubernetes as a master node to your machine."
echo -e "This script only works on Debian based machines"
echo -e "It will ask for your input during installation for installation paths"
echo -e "This will make substantial changes to your system. Please view the contents of the script before installation."
read -n 1 -s -r -p "Press enter to start install"
echo

stage="${green}Pre-flight configuration$reset"

echo -e "${stage} >> Upgrading system"
if ! command -v apt &>/dev/null;
then
  echo -e "${red}System is not based on Debian nor Ubuntu. Abort"
  exit 1
fi
apt update
apt upgrade -y
if [ -f /var/run/reboot-required ]; then
  echo -e "Reboot required after update. Please reboot the server and rerun this script"
  exit 2
fi

echo -e "${stage} >> Configuring firewall"
echo
echo -e "${stage} >> Configuring iptables. Answer yes on the next message box"
sleep 3s
apt install iptables-persistent -y
if command -v ufw &>/dev/null
then
    echo -e "${stage} >> Disabling UFW"
    ufw disable
fi
iptables -P FORWARD ACCEPT
echo
echo -e "${stage} >> Saving firewall configuration. Answer yes to both questions in 3 seconds"
sleep 3s
dpkg-reconfigure iptables-persistent

echo -e "${stage} >> Disabling swap"
swapoff -a
sed -i '/swap/d' /etc/fstabsed -i 's,\(/.*[[:space:]]none[[:space:]]*swap[[:space:]]\),#\1,' /etc/fstab

if command -v setenforce &>/dev/null
then
    echo -e "${stage} >> Disabling SELinux"
    setenforce 0
    sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config
fi

echo -e "${stage} >> Configuring network"
cat >> /etc/sysctl.d/kubernetes.conf<<EOF
net.bridge.bridge-nf-call-ip6tables=1
net.bridge.bridge-nf-call-iptables=1
EOF
sysctl --system

stage="${green}Installing Docker$reset"

echo -e "${stage} >> Remove old versions if they exist"
apt remove docker docker-engine docker.io containerd runc -y

echo -e "${stage} >> Installing Docker"
VERSION=19.03
curl https://get.docker.com/ | sh
apt update
apt install docker-compose -y

echo -e "${stage} >> Configuring Docker"
cat >> /etc/docker/daemon.json<<EOF
{
    "exec-opts": ["native:cgroupdriver=systemd"],
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "100m"
    },
    "storage-driver": "overlay2"
}
EOF
systemctl restart docker
usermod -aG docker $SUDO_USER

stage="${green}Installing Kubernetes$reset"

echo -e "${stage} >> Preparing sources"
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
apt update
apt install kubeadm=1.18.12-00 kubectl=1.18.12-00 kubelet=1.18.12-00 -y
apt-mark hold kubeadm kubelet kubectl docker-ce

echo -e "${stage} >> Configuring kubeadm"
USER_HOME=$(eval echo ~$SUDO_USER)
KUBEADM_CONFIG=$USER_HOME/kubeadm-config.yaml
cat > $KUBEADM_CONFIG <<EOF
kind: ClusterConfiguration
apiVersion: kubeadm.k8s.io/v1beta2
kubernetesVersion: v1.18.12
networking:
  podSubnet: 10.244.0.1/16
---
kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
cgroupDriver: systemd
EOF

echo "Configuration file written to $KUBEADM_CONFIG"

echo -e "${stage} >> Initializing cluster"
kubeadm init --config $KUBEADM_CONFIG

echo -e "${stage} >> Setting up container network interfaces"
kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml

echo -e "${stage} >> Copy Kubernetes config to home directory"
mkdir -p $USER_HOME/.kube
cp -i /etc/kubernetes/admin.conf $USER_HOME/.kube/config
chown $(id -u $SUDO_USER):$(id -g $SUDO_USER) $USER_HOME/.kube/config
# Root also needs a copy of the config to access kubectl
mkdir -p /root/.kube
cp -i /etc/kubernetes/admin.conf /root/.kube/config

echo -e "${stage} >> Configuring Kubernetes"
kubectl taint nodes $HOSTNAME node-role.kubernetes.io/master-

stage="${green}Setting up NFS server$reset"
echo -e "$stage >> Installing NFS"
apt install nfs-kernel-server -y
echo -e "$stage >> Creating NFS share directory"

path_valid=false
nfs_path="/opt/nfs"
while [ $path_valid = false ]; do
    echo -e "$stage >> Please enter the path of your NFS share. Make sure the path is accessible [default: /opt/nfs]"
    read -r -p "NFS share path: " nfs_path
    if [ -z $nfs_path ]
    then
        nfs_path="/opt/nfs"
        path_valid=true
    elif [ ! -d $nfs_path ]
    then
        if mkdir -p $nfs_path ; then
            path_valid=true
        else
            echo -e "Path $nfs_path cannot be created"
        fi
    else
        path_valid=true
    fi
done

mkdir -p "$nfs_path/registry-storage"
chown nobody:nogroup "$nfs_path/registry-storage"
chmod 1777 "$nfs_path/registry-storage"

echo -e "$stage >> Configuring NFS share"
cat >> /etc/exports<<EOF
$nfs_path/registry-storage *(rw,sync,no_subtree_check,root_squash)
EOF

exportfs -a
systemctl enable nfs-kernel-server
systemctl restart nfs-kernel-server

stage="${green}Setting up monitoring services$reset"
echo -e "$stage >> Installing Prometheus"
git clone https://github.com/bibinwilson/kubernetes-prometheus /tmp/kubernetes-prometheus
kubectl create ns monitoring
kubectl apply -f /tmp/kubernetes-prometheus/prometheus-deployment.yaml
kubectl apply -f /tmp/kubernetes-prometheus/prometheus-service.yaml
kubectl apply -f /tmp/kubernetes-prometheus/config-map.yaml
kubectl apply -f /tmp/kubernetes-prometheus/clusterRole.yaml

echo -e "$stage >> Installing kube-state-metrics"
git clone https://github.com/devopscube/kube-state-metrics-configs /tmp/kube-state-metrics-configs
kubectl apply -f /tmp/kube-state-metrics-configs/

echo -e "$stage >> Installing Node Exporter"
git clone https://github.com/bibinwilson/kubernetes-node-exporter /tmp/kubernetes-node-exporter
kubectl apply -f /tmp/kubernetes-node-exporter/

echo -e "$stage >> Waiting for Prometheus"
max_attempt=60
eta=0
until $(curl --output /dev/null --silent --head http://localhost:30000); do
  printf '.'
  if [ $eta -eq $max_attempt ]; then
    printf 'failed\n'
    echo -e "$stage >> ${red}Prometheus is still unreachable. Please debug the issue manually later${reset}"
    sleep 10s
    break
  fi
  sleep 1s
  eta=$(($eta+1))
done

stage="${green}Setting up registry$reset"

echo -e "$stage >> Deploying registry"
cat > /tmp/registry.yml << EOF
# Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: registry
spec:
  replicas: 1
  selector:
    matchLabels:
      app: registry
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: registry
    spec:
      containers:
        - name: registry
          image: registry:2
          ports:
            - containerPort: 5000
          volumeMounts:
            - name: registry-storage
              mountPath: /var/lib/registry
              readOnly: false
      volumes:
        # Registry storage
        - name: registry-storage
          persistentVolumeClaim:
            claimName: registry-storage-claim
---
# Service
apiVersion: v1
kind: Service
metadata:
  labels:
    app: registry
  name: registry-service
spec:
  selector:
    app: registry
  type: NodePort
  ports:
    - name: registry-port
      port: 5000
      targetPort: 5000
      nodePort: 32000
      protocol: TCP
---
# PersistentVolumeClaim. The reason to use a PersistentVolumeClaim is the volume will stay even after the pods lifetime.
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: registry-storage-claim
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
EOF
kubectl apply -f /tmp/registry.yml
cat > /tmp/registry-storage.yml << EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: registry-volume
spec:
  capacity:
    storage: 20Gi
  accessModes:
    - ReadWriteOnce
  nfs:
    path: $nfs_path/registry-storage
    server: $HOSTNAME
    readOnly: false
EOF
kubectl apply -f /tmp/registry-storage.yml

stage="${green}Clean up$reset"
echo -e "$stage >> Deleting temporary files"
# Revoke permissions
rm -rf /root/.kube


echo -e "${green}Done! Reboot is recommended to ensure all of the services are properly started up"
echo -e "Here are the following services deployed and its respective port number"
echo -e "Prometheus server: 30000"
echo -e "Registry server: 32000"
echo -e "NFS server: 2049"
echo -e
interface=$(ip route | grep default | sed -e "s/^.*dev.//" -e "s/.proto.*//")
ip=$(ip -f inet addr show $interface | sed -En -e 's/.*inet ([0-9.]+).*/\1/p')
echo -e "Your Prometheus API endpoint: http://$ip:30000/api/v1"
echo -e "Your Kubernetes configuration file is at $USER_HOME/.kube/config"
echo -e
echo -e "To push images to the registry, tag and push your Docker container like so"
echo -e "docker tag IMAGE_NAME $ip:32000/IMAGE_NAME:latest"
echo -e "docker push $ip:32000/IMAGE_NAME:latest"
echo -e
echo -e "To add more nodes to the cluster, please refer to the documentation"
echo -e
echo -e "Setup took $(($SECONDS / 60)) minutes and $(($SECONDS % 60)) seconds$reset"