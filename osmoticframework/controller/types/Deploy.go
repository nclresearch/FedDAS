package types

//Deployment information for Docker

type DeployArgs struct {
	//The image of the container.
	//[image]:[version]
	Image string
	//Exposed ports. Must be TCP, UDP, or SCTP
	ExposePorts []ExposePort
	//Arguments and environment variables

	//The entry point of the container. This states the file that first runs when a container starts
	//Leave empty for image default
	Entrypoint []string
	//The command of the container. If the entrypoint is not supplied, this would be the entrypoint. Otherwise, this is treated as command line arguments
	//Leave empty for image default
	Command []string
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
	Volumes      []Volume

	//GPU support - leave nil if not needed
	GPU *GPU
	//Restart policy
	RestartPolicy RestartPolicy
	//Device mounting support
	Devices []Device
	//Pull options
	PullOptions PullOption
}

// GPU support
type GPU struct {
	// Use only ONE of the following options. If you specify both, only the GPU count will be used.

	// -1 for automatically detecting all available GPUs.
	Count *int64
	// Or list of device IDs that is recognizable by the device driver (For Nvidia GPUs, check nvidia-smi. This should be an integer from 0 to n, where (n - 1) is the number of GPUs you have)
	// Convert the numeric ID to a string for this parameter
	DeviceIDs []string
}

//Represents a device file (e.g. Anything in /dev) on the edge device
type Device struct {
	HostDevicePath      string
	ContainerDevicePath string
	//This controls what can be performed on the device. They are:
	//Read (r)
	//Write (w)
	//Mknod (m)
	//By default all three permissions are allowed (rwm)
	CgroupPermissions string
}

//Volume information
//In Docker, this mounts a container directory to the host directory
type Volume struct {
	ContainerPath string
	HostPath      string
	ReadOnly      bool
}

//An expose port to the container
type ExposePort struct {
	HostPort      uint16
	ContainerPort uint16
	Protocol      Protocol
}

//Authentication information for pulling images from Docker Hub
type AuthInfo struct {
	Username string
	Password string
}

//Struct for one environment value
type Environment struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type RestartPolicy string

const (
	//Always will not be available.
	//As agents may disconnect from the controller and any dangling containers running even after a system restart. It will be a waste of resources.
	RestartOnFailure     RestartPolicy = "on-failure"
	RestartUnlessStopped RestartPolicy = "unless-stopped"
	RestartNever         RestartPolicy = "no"
)

type PullOption string

const (
	PullAlways     PullOption = "always"
	PullIfNotExist PullOption = "ifNotExist"
)

//Container struct. Used when a listing container call is made
type Container struct {
	ID    string
	Image string
	//Container arguments
	Command string
	Status  string
	//Total container file size
	SizeRootFs int64
	//Container file size, excluding its base image
	SizeRw        int64
	Volumes       []Volume
	ExposePorts   []ExposePort
	Environments  []Environment
	Devices       []Device
	Entrypoint    []string
	RestartPolicy RestartPolicy
	MemLimit      int64
	MemSoftLimit  int64
	GPU           string
}

type Protocol string

const (
	TCP  Protocol = "tcp"
	UDP  Protocol = "udp"
	SCTP Protocol = "sctp"
)
