package docker

import (
	"context"
	"errors"
	"fmt"
	"math"
	"osmoticframework/agent/log"
	"osmoticframework/agent/types"
	"regexp"
	"strconv"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

//The deployer interfaces with Docker to do the actual deploying task

var client *docker.Client

//Initialize the docker client
//You can communicate with docker through REST or directly through the host environment
//To maintain low latency, the agent only supports running it on the edge device directly.
func Init() error {
	cli, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	client = cli
	return nil
}

func Run(spec types.DeployArgs, auth types.AuthInfo) (string, error) {
	//Check if image exist in host.
	if spec.PullOptions == "" || spec.PullOptions == types.PullIfNotExist {
		imageExist, err := isImageExist(spec.Image)
		if err != nil {
			return "", err
		}
		//Pull the image from Docker Hub
		if !imageExist {
			err = pullImage(spec.Image, auth)
		}
		if err != nil {
			return "", err
		}
	} else if spec.PullOptions == types.PullAlways {
		err := pullImage(spec.Image, auth)
		if err != nil {
			return "", err
		}
	}

	log.Info.Println("Deploying " + spec.Image + " using SDK")
	//Setting up container configuration
	exposePorts := make(map[docker.Port]struct{})
	portBinding := make(map[docker.Port][]docker.PortBinding)
	for _, port := range spec.ExposePorts {
		containerPort := strconv.FormatInt(int64(port.ContainerPort), 10)
		hostPort := strconv.FormatInt(int64(port.HostPort), 10)
		var protocol types.Protocol
		if port.Protocol == "" {
			protocol = types.TCP
		} else {
			protocol = port.Protocol
		}
		port := docker.Port(containerPort + "/" + string(protocol))
		exposePorts[port] = struct{}{}
		var binding []docker.PortBinding
		binding = append(binding, docker.PortBinding{
			HostIP:   "0.0.0.0",
			HostPort: hostPort,
		})
		portBinding[port] = binding
	}

	//Environment variables goes like this
	//KEY=VALUE
	var env []string
	for _, v := range spec.Environment {
		env = append(env, v.Name+"="+v.Value)
	}

	//Create docker configuration.
	var config docker.Config
	config = docker.Config{
		ExposedPorts: exposePorts,
		Env:          env,
		Image:        spec.Image,
	}

	if len(spec.Command) != 0 {
		config.Cmd = spec.Command
	}

	if len(spec.Entrypoint) != 0 {
		config.Entrypoint = spec.Entrypoint
	}

	//Volume information
	//Only supports host path volume mounting on edge devices
	//You mount volumes in Docker like this
	// hostPath:mountPoint:readWriteFlag
	var mounts = make([]string, 0)
	for _, volume := range spec.Volumes {
		var mountFlag string
		if volume.ReadOnly {
			mountFlag = "ro"
		} else {
			mountFlag = "rw"
		}
		mounts = append(mounts, fmt.Sprintf("%s:%s:%s", volume.HostPath, volume.ContainerPath, mountFlag))
	}

	//Network configuration
	hostConfig := docker.HostConfig{
		PortBindings:  portBinding,
		NetworkMode:   "host",
		Binds:         mounts,
		RestartPolicy: docker.NeverRestart(),
	}
	if spec.MemLimit != 0 {
		hostConfig.Memory = spec.MemLimit
	}
	if spec.MemSoftLimit != 0 {
		hostConfig.MemoryReservation = spec.MemSoftLimit
	}
	if spec.GPU != nil {
		if spec.GPU.Count != nil {
			hostConfig.DeviceRequests = []docker.DeviceRequest{
				{
					Driver:       "nvidia",
					Count:        int(*spec.GPU.Count),
					Capabilities: [][]string{{"gpu"}},
				},
			}
		}
		if spec.GPU.DeviceIDs != nil {
			hostConfig.DeviceRequests = []docker.DeviceRequest{
				{
					Driver:       "nvidia",
					DeviceIDs:    spec.GPU.DeviceIDs,
					Capabilities: [][]string{{"gpu"}},
				},
			}
		}
	}
	//Device files
	if spec.Devices != nil && len(spec.Devices) != 0 {
		for _, device := range spec.Devices {
			dockerDevice := docker.Device{
				PathOnHost:      device.HostDevicePath,
				PathInContainer: device.ContainerDevicePath,
			}
			if device.ContainerDevicePath != "" {
				dockerDevice.CgroupPermissions = device.CgroupPermissions
			}
			hostConfig.Devices = append(hostConfig.Devices, dockerDevice)
		}
	}
	if spec.RestartPolicy != "" {
		switch spec.RestartPolicy {
		case types.RestartNever:
			hostConfig.RestartPolicy = docker.NeverRestart()
		case types.RestartOnFailure:
			//Restart endlessly
			hostConfig.RestartPolicy = docker.RestartOnFailure(math.MaxInt32)
		case types.RestartUnlessStopped:
			hostConfig.RestartPolicy = docker.RestartUnlessStopped()
		default:
			//Fallback
			hostConfig.RestartPolicy = docker.NeverRestart()
		}
	}
	networkConfig := docker.NetworkingConfig{EndpointsConfig: nil}
	options := docker.CreateContainerOptions{
		Config:           &config,
		HostConfig:       &hostConfig,
		NetworkingConfig: &networkConfig,
		Context:          context.Background(),
	}

	//Create the container
	newContainer, err := client.CreateContainer(options)
	if err != nil {
		return "", err
	}

	//Start the container
	err = client.StartContainer(newContainer.ID, newContainer.HostConfig)
	if err != nil {
		return "", err
	}

	log.Info.Println("Deployed container " + newContainer.ID)
	log.Info.Println("Image: " + spec.Image)
	if spec.ExposePorts != nil {
		log.Info.Println("Ports exposed: ")
		for _, port := range spec.ExposePorts {
			log.Info.Printf("  - %d:%d/%s\n", port.HostPort, port.ContainerPort, port.Protocol)
		}
	}
	log.Info.Printf("Command arguments: %#v\n", spec.Command)
	log.Info.Printf("Entrypoint: %#v\n", spec.Entrypoint)
	if spec.MemLimit != 0 {
		log.Info.Printf("Memory hard limit: %d\n", spec.MemLimit)
	}
	if spec.MemSoftLimit != 0 {
		log.Info.Printf("Memory soft limit: %d\n", spec.MemSoftLimit)
	}
	if spec.Environment != nil {
		log.Info.Println("Environment variables:")
		for _, env := range spec.Environment {
			log.Info.Printf("  - %s:%s\n", env.Name, env.Value)
		}
	}
	if spec.Volumes != nil {
		log.Info.Println("Mounted volumes")
		for _, volume := range spec.Volumes {
			log.Info.Printf("  - ro:%t %s:%s\n", volume.ReadOnly, volume.HostPath, volume.ContainerPath)
		}
	}
	if spec.Devices != nil {
		log.Info.Println("Mounted devices")
		for _, device := range spec.Devices {
			log.Info.Printf("  - %s %s:%s\n", device.CgroupPermissions, device.HostDevicePath, device.ContainerDevicePath)
		}
	}
	if spec.GPU != nil {
		log.Info.Println("GPU support enabled")
		if spec.GPU.Count != nil {
			log.Info.Printf("  - Count: %d\n", *spec.GPU.Count)
		}
		if spec.GPU.DeviceIDs != nil {
			log.Info.Printf("  - Device IDs: %v\n", spec.GPU.DeviceIDs)
		}
	}
	if spec.RestartPolicy != "" {
		log.Info.Println("Restart policy: " + spec.RestartPolicy)
	}
	return newContainer.ID, nil
}

func Stop(containerId string) error {
	err := client.StopContainer(containerId, 60)
	return err
}

func Delete(containerId string, deleteImage bool) error {
	container, err := client.InspectContainerWithOptions(docker.InspectContainerOptions{
		Context: context.Background(),
		ID:      containerId,
	})
	if err != nil {
		return err
	}
	imageTag := container.Image
	options := docker.RemoveContainerOptions{
		ID:            containerId,
		RemoveVolumes: true,
		Force:         false,
		Context:       context.Background(),
	}
	err = client.RemoveContainer(options)
	if err != nil {
		return err
	}
	if deleteImage {
		log.Info.Println("Deleting image " + imageTag)
		err = client.RemoveImage(imageTag)
		if err != nil {
			return err
		}
	}
	return nil
}

func Update(containerId string, spec types.DeployArgs, auth types.AuthInfo) (string, error) {
	err := Stop(containerId)
	if err != nil {
		return "", err
	}
	err = Delete(containerId, false)
	if err != nil {
		return "", err
	}
	newContainerId, err := Run(spec, auth)
	if err != nil {
		return "", err
	}
	return newContainerId, nil
}

//List all containers in greater detail, running or not
func ListDetailed() ([]types.Container, error) {
	dContainers, err := client.ListContainers(docker.ListContainersOptions{
		All:     true,
		Context: context.Background(),
	})
	if err != nil {
		return nil, err
	}
	containerList := make([]types.Container, 0)
	for _, dContainer := range dContainers {
		container, err := Inspect(dContainer.ID)
		if err != nil {
			return nil, err
		}
		containerList = append(containerList, *container)
	}
	return containerList, nil
}

//List all containers by calling the Docker API.
func DockerList() ([]docker.APIContainers, error) {
	return client.ListContainers(docker.ListContainersOptions{
		Context: context.Background(),
		All:     true,
	})
}

//Inspects a container
func Inspect(containerId string) (*types.Container, error) {
	//Inspect container for details
	iContainer, err := dockerInspect(containerId)
	if err != nil {
		return nil, err
	}
	//Basic information
	container := types.Container{
		ID:         iContainer.ID,
		Image:      iContainer.Config.Image,
		Command:    strings.Join(iContainer.Config.Cmd, " "),
		Status:     iContainer.State.Status,
		SizeRootFs: iContainer.SizeRootFs,
		SizeRw:     iContainer.SizeRw,
	}
	//Volumes
	volumes := make([]types.Volume, 0)
	for _, dVolume := range iContainer.Mounts {
		volumes = append(volumes, types.Volume{
			ContainerPath: dVolume.Destination,
			HostPath:      dVolume.Source,
			ReadOnly:      dVolume.RW,
		})
	}
	//Exposed ports
	container.Volumes = volumes
	exposePorts := make([]types.ExposePort, 0)
	for apiContainerBinding, apiHostBinding := range iContainer.HostConfig.PortBindings {
		exposePorts = append(exposePorts, types.ExposePort{
			HostPort:      parsePort(apiHostBinding[0].HostPort),
			ContainerPort: parsePort(apiContainerBinding.Port()),
			Protocol:      types.Protocol(apiContainerBinding.Proto()),
		})
	}
	container.ExposePorts = exposePorts
	//Environment variables
	environments := make([]types.Environment, 0)
	for _, dEnvironment := range iContainer.Config.Env {
		environments = append(environments, types.Environment{
			Name:  strings.SplitN(dEnvironment, "=", 2)[0],
			Value: strings.SplitN(dEnvironment, "=", 2)[1],
		})
	}
	container.Environments = environments
	//Host bind devices
	devices := make([]types.Device, 0)
	for _, dDevices := range iContainer.HostConfig.Devices {
		devices = append(devices, types.Device{
			HostDevicePath:      dDevices.PathOnHost,
			ContainerDevicePath: dDevices.PathInContainer,
			CgroupPermissions:   dDevices.CgroupPermissions,
		})
	}
	container.Devices = devices
	//Entrypoint
	container.Entrypoint = iContainer.Config.Entrypoint
	//Restart policy
	container.RestartPolicy = types.RestartPolicy(iContainer.HostConfig.RestartPolicy.Name)
	//Resource quotas
	if iContainer.HostConfig.Memory == 0 {
		//Memory limit unset
		container.MemLimit = -1
	} else {
		container.MemLimit = iContainer.HostConfig.Memory
	}
	if iContainer.HostConfig.MemoryReservation == 0 {
		//Memory reservation limit unset
		container.MemSoftLimit = -1
	} else {
		container.MemSoftLimit = iContainer.HostConfig.MemoryReservation
	}
	//GPU
	var gpuOption string
	for _, deviceReq := range iContainer.HostConfig.DeviceRequests {
		//Search for the string GPU in capabilities
		for _, capability := range deviceReq.Capabilities {
			for _, subCap := range capability {
				if subCap == "gpu" {
					//Search for device ID
					if deviceReq.DeviceIDs == nil {
						gpuOption = "all"
					} else {
						gpuOption = "device="
						gpuOption = strings.Join(deviceReq.DeviceIDs, ",")
					}
				}
			}
		}
	}
	if gpuOption != "" {
		container.GPU = gpuOption
	}
	return &container, nil
}

func parsePort(portString string) uint16 {
	port, _ := strconv.ParseUint(portString, 10, 16)
	return uint16(port)
}

//Inspects the container by calling the Docker API
func dockerInspect(containerID string) (*docker.Container, error) {
	return client.InspectContainerWithOptions(docker.InspectContainerOptions{
		Context: context.Background(),
		ID:      containerID,
	})
}

func RegisterListener(options docker.EventsOptions, listener chan *docker.APIEvents) error {
	return client.AddEventListenerWithOptions(options, listener)
}

func DeregisterListener(listener chan *docker.APIEvents) error {
	return client.RemoveEventListener(listener)
}

//Checks if the edge device has the image available locally
func isImageExist(imageName string) (bool, error) {
	images, err := client.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		return false, err
	}
	for _, image := range images {
		if image.RepoTags == nil {
			continue
		}
		if strings.Contains(image.RepoTags[0], imageName) {
			return true, nil
		}
	}
	return false, nil
}

//Pulls the image, by default from Docker Hub.
//Does not support pulling by digest! Tags only.
//Empty tags will default to latest.
//You can specify the image name to force Docker to pull from a separate registry
func pullImage(imageTag string, info types.AuthInfo) error {
	log.Info.Println("Pulling image " + imageTag)
	// This regex is copied and modified from Docker's source code
	// See https://github.com/distribution/distribution/blob/main/reference/regexp.go
	re := regexp.MustCompile(`((?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])(?:(?:\.(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]))+)?(?::[0-9]+)?/)?[a-z0-9]+(?:(?:(?:[._]|__|[-]*)[a-z0-9]+)+)?(?:(?:/[a-z0-9]+(?:(?:(?:[._]|__|[-]*)[a-z0-9]+)+)?)+)?)(:([\w-]+))?`)
	matches := re.FindAllStringSubmatch(imageTag, -1)
	if len(matches) == 0 {
		return errors.New("invalid image tag")
	}
	image := matches[0][1]
	tag := "latest"
	if len(matches[0]) == 3 {
		tag = matches[0][3]
	}
	ctx := context.Background()
	var err error
	if info == (types.AuthInfo{}) {
		err = client.PullImage(docker.PullImageOptions{
			Context:    ctx,
			Repository: image,
			Tag:        tag,
		}, docker.AuthConfiguration{})
	} else {
		err = client.PullImage(docker.PullImageOptions{
			Repository: image,
			Tag:        tag,
			Context:    ctx,
		}, docker.AuthConfiguration{
			Username: info.Username,
			Password: info.Password,
		})
	}
	return err
}
