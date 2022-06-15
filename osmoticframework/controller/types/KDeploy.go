package types

import (
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

//Deployment information for Kubernetes

//Kubernetes Pod arguments
type KPodArgs struct {
	//The image of the container.
	//[image]:[version]
	Image string
	//The label is used in selectors, where services pick pods based on the label.
	Label map[string]string
	//Name of the container
	Name string
	//Exposed ports. Must be TCP, UDP, or SCTP
	ExposePorts []ExposePort
	//The entrypoint of the container. This is equivalent to the ENTRYPOINT variable in Dockerfile
	//Defaults to container original configuration if left nil
	//Not executed in a shell
	Entrypoint []string
	//The arguments of the entrypoint. This is equivalent to the CMD variable in Dockerfile
	//Defaults to container original configuration if left nil
	//If the entrypoint is not specified in the container, the arguments will act as the entrypoint.
	//Not executed in a shell
	Arguments []string
	/*
		Environment variables
		Each entry is a key value string pair.
		Example:
		- Name: "KAFKA_BROKER_ID",
		  Value: "0",
	*/
	Environment []Environment
	//Memory limit for the container (in bytes)
	/*
		There are 2 limits.
		The MemLimit refers to the hard limit. The container cannot use more than this number.
		The MemSoftLimit refers to the memory the container guaranteed to get. If the container uses more than that and later free someone, the kernel will claim that memory back.
	*/
	MemLimit     int64
	MemSoftLimit int64
	//Number of whole CPUs
	//Soft limit of CPU means the container is guaranteed to have this much resources on the CPU
	//See https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu
	CPULimit     float64
	CPUSoftLimit float64
	//Volumes
	Volumes []KVolume
	//Name of the accelerator
	//The name is defined in the cluster via kubectl
	Accelerator string
	//Scheduling the number of whole GPUs. This is brand specific.
	//For more details see https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/
	Nvidia int64
	AMD    int64
	//Image pulling policy.
	//This controls if the image should be pulled from Docker Hub and registry or refer to local cache only
	PullPolicy corev1.PullPolicy
	//Restart policy
	//This option has no effect on deployments (They always automatically restart). It will be locked as "Always"
	//For jobs and cronjobs, the RestartPolicy must always be either "OnFailure" or "Never"
	RestartPolicy corev1.RestartPolicy
}

//Kubernetes deployment arguments
type KDeployArgs struct {
	//Pod specification
	PodArgs KPodArgs
	//The deployment name
	DeploymentName string
	//Number of replicas
	Replicas int32
	//Deploy strategy
	Strategy v1.DeploymentStrategyType
}

//Kubernetes job arguments
//This runs a container once. You cannot restart it after it finishes execution (unless it crashed)
//If the cluster has TTLAfterFinish feature enabled (alpha feature), the job will be deleted 30 seconds after it's completed
type KJobArgs struct {
	//Pod specification
	//Caution: Your pod labels must be unique from other jobs and deployments. Otherwise, Kubernetes cannot determine if the job has finished
	PodArgs KPodArgs
	//The job name
	JobName string
	//Back off limit. The amount of times this container is allowed to restart
	BackOffLimit int32
}

//Kubernetes cronjob arguments
//This runs a job periodically according to crontab arguments.
type KCronjobArgs struct {
	//Job specification
	//Caution: Your pod labels must be unique from other jobs and deployments. Otherwise, Kubernetes cannot determine if the job has finished
	JobArgs KJobArgs
	//The cronjob name
	JobName string
	//Cron schedule string
	//This uses the same syntax as crontab
	Cron string
}

//Kubernetes volume
type KVolume struct {
	//Where to mount the volume in the container
	MountPoint string
	ReadOnly   bool
	Type       VolumeType
	//There are many volume types in Kubernetes.
	//Only use ONE of them when deploying
	//Supported volume types on the orchestrator are: NFS, AWS EBS volumes and hostPath (deprecated)

	//Host directory
	HostPath *HostPath
	//NFS volume
	NFS *NFS
	//AWS EBS volume. You must first create a EBS volume in AWS
	AwsEbs *AwsEbs
	//EmptyDir volume
	EmptyDir *EmptyDir
	//ConfigMap volume
	ConfigMap *ConfigMapVolume
}

type VolumeType string

const (
	TypeHostPath  VolumeType = "hostpath"
	TypeNFS       VolumeType = "nfs"
	TypeAWSEBS    VolumeType = "awsebs"
	TypeEmptyDir  VolumeType = "emptyDir"
	TypeConfigMap VolumeType = "configMap"
)

//Host path mounting
//Deprecated : Use NFS instead as hostpath does not support multi-node
type HostPath struct {
	Path      string
	MountType corev1.HostPathType
	//Specify which host this volume will mount to using its hostname defined in Kubernetes
	//This will force the pod to deploy on that host
	//You can find the names of nodes via `kubectl get nodes`
	NodeHostname string
}

//NFS server
//The pod will mount to a path in an NFS server as a volume
type NFS struct {
	Server string
	Path   string
}

//AWS EBS volume
//You must first create an EBS volume on AWS
//This mounts the entire EBS disk to the target pod. You can mitigate this by partitioning the disk and provision it to multiple pods
type AwsEbs struct {
	//Volume ID of the EBS volume
	VolumeID string
	//File system type. Such as ext4
	FSType    string
	Partition int32
}

//Empty directory
//Creates an empty directory as a volume. This volume persists until the pod is removed or if the cluster is shut down
//Can be useful for pods to store data and persist through crashes
type EmptyDir struct {
	//Where to store this emptyDir
	//Default goes to disk
	//Memory goes to tmpfs (In memory, be careful of memory usage. Gets cleared when server shuts down)
	//Huge pages (In virtual memory. No size limits. Gets cleared when server shuts down)
	StorageMedium corev1.StorageMedium
	//Size limit in bytes - nil for unlimited
	SizeLimit *int64
}

//ConfigMap volume
//A ConfigMap is a key-value map that is stored in Kubernetes
//In most cases, this is a static file such as configurations files that you want to be available to your pods
//To mount a ConfigMap, you must first create a ConfigMap in Kubernetes
type ConfigMapVolume struct {
	//Name of the ConfigMap
	Name string
	//Permissions to apply to the volume
	Mode *int32
}

//A Kubernetes service
type KServiceArgs struct {
	//The service name
	Name string
	//Exposed ports. Must be TCP, UDP, or SCTP
	Ports []KServicePort
	//The selector label
	//Used to select which deployment to expose.
	//For example:
	//A deployment has the key-value { "app" : "web" } in their label
	//The service selector then also must contain all of those labels to expose those deployment.
	Selector map[string]string
	//The current version of this service
	//Needed when updating a service
	//Obtained after deploying.
	ResourceVersion string
	//Service type
	Type corev1.ServiceType
	//Cluster IP
	ClusterIP string
	//External DNS name
	//Only has effect when the type of service is ExternalName
	//Maps a service selector to a local DNS name. This does not map your service to a public DNS domain.
	ExternalName string
}

//An exposed port of a service
type KServicePort struct {
	//Name of the exposed port
	//Needed when updating a service. Can be omitted empty (If the service doesn't need to be updated, ever).
	Name string
	//The port to be exposed to internally. Pods can connect to the service through the cluster IP.
	//The cluster IP can be obtained in CLI through `kubectl get service`
	Host int32
	//The port to be exposed from (the container/pod)
	TargetPort int32
	//The port to be exposed to externally. Outside clients can connect to the service through this port and node IP.
	//Pods can also connect to the service through this way as well as the cluster IP method.
	//The node IP can be obtained in CLI through `kubectl describe node [nodeIP/minikube]`
	NodePort int32
	//Protocol to be used
	Protocol Protocol
}

//Kubernetes config map. Config maps are static files that can be mounted to pods as configurations/secrets
//As of Kubernetes v1.18, ConfigMaps act the same as secrets (Secrets are only base64 encoded, not encrypted).
//Config maps have a hard size limit of 1MB that cannot be changed
type KConfigMap struct {
	//Name of the config map
	Name string
	//Data to be stored in the config map
	//The key is the name of the file
	//The value is the content of the file
	Data map[string]string
	//Binary data to be stored in the config map
	//If you need to store large binaries, it's better to have it built into the image or download from a remote source
	//The key is the name of the file
	//The value is the content of the file
	BinaryData map[string][]byte
}
