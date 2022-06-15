package docker

import (
	"os/exec"
	"osmoticframework/agent/types"
	"osmoticframework/controller/util"
	"regexp"
	"strings"
	"testing"
	"time"
)

var currentContainerId string

//Due to how the deployer works. These tests must be done sequentially
func TestDeployer(t *testing.T) {
	err := Init()
	if err != nil {
		t.Error("Docker not installed")
	}
	containers := make([]string, 0)
	deploy := types.DeployArgs{
		Image: "ubuntu:20.04",
		ExposePorts: []types.ExposePort{
			{
				HostPort:      1234,
				ContainerPort: 1234,
				Protocol:      types.TCP,
			},
		},
		Volumes: []types.Volume{
			{
				HostPath:      "/",
				ContainerPath: "/rootfs",
				ReadOnly:      true,
			},
		},
		Entrypoint:   []string{"bash"},
		Command:      []string{"bash"},
		MemLimit:     512000000,
		MemSoftLimit: 256000000,
		Environment: []types.Environment{
			{
				Name:  "test",
				Value: "this is a test",
			},
		},
		RestartPolicy: types.RestartOnFailure,
		Devices: []types.Device{
			{
				HostDevicePath:      "/dev/loop0",
				ContainerDevicePath: "/dev/loop0",
				CgroupPermissions:   "r",
			},
		},
	}
	t.Run("Run", func(t *testing.T) {
		currentContainerId = runTest(deploy, t)
	})
	containers = append(containers, currentContainerId)
	t.Run("Inspect", func(t *testing.T) {
		inspectTest(currentContainerId, deploy, t)
	})
	if !t.Failed() {
		t.Run("List", func(t *testing.T) {
			listTest(currentContainerId, deploy, t)
		})
		t.Run("Update", func(t *testing.T) {
			currentContainerId = updateTest(currentContainerId, deploy, t)
		})
		containers[0] = currentContainerId
	} else {
		t.Log("[SKIP] Run/Inspect test failed. Skipping update container test")
	}
	t.Run("GPU run", func(t *testing.T) {
		_, err = exec.LookPath("nvidia-smi")
		if err != nil {
			t.Skip("[SKIP] No Nvidia drivers found. Skipping GPU test")
		} else {
			currentContainerId = runGPUTest(t)
			containers = append(containers, currentContainerId)
		}
	})
	// Padding time to allow the container to run first
	time.Sleep(time.Second * 5)
	t.Run("Stop", func(t *testing.T) {
		stopTest(containers, t)
	})
	t.Run("Delete", func(t *testing.T) {
		deleteTest(containers, t)
	})
}

func TestPullImage(t *testing.T) {
	err := pullImage("hello-world", types.AuthInfo{})
	if err != nil {
		t.Error(err)
	}
	err = pullImage("hello-world:latest", types.AuthInfo{})
	if err != nil {
		t.Error(err)
	}
	err = pullImage("gcr.io/google.com/cloudsdktool/google-cloud-cli:latest", types.AuthInfo{})
	if err != nil {
		t.Error(err)
	}
	err = pullImage("gcr.io:443/google.com/cloudsdktool/google-cloud-cli:latest", types.AuthInfo{})
	if err != nil {
		t.Error(err)
	}
}

func TestInvalidDeploy(t *testing.T) {
	deploy := types.DeployArgs{
		Image: "ubuntu",
		Devices: []types.Device{
			{
				HostDevicePath:      "/dev/sdfhjkdsbn",
				ContainerDevicePath: "/dev/sda1",
			},
		},
	}
	_, err := Run(deploy, types.AuthInfo{})
	if err == nil {
		t.Error("This should error but it did not")
	}
}

func TestIsImageExist(t *testing.T) {
	exist, err := isImageExist("fgdjhfghjkl")
	if err != nil {
		t.Error("Failed to execute")
	}
	t.Log(exist)
	exist, err = isImageExist("ubuntu")
	if err != nil {
		t.Error("Failed to execute")
	}
	t.Log(exist)
}

func runTest(spec types.DeployArgs, t *testing.T) string {
	re := regexp.MustCompile("[a-f0-9]{64}")
	//The logger will not work as the paths are not initialized in the test.
	containerId, err := Run(spec, types.AuthInfo{})
	if err != nil {
		t.Error(err)
	}
	if !re.MatchString(containerId) {
		t.Errorf("Container not match regex. Got %s, Want format [a-f0-9]{64}", containerId)
	}
	return containerId
}

func runGPUTest(t *testing.T) string {
	re := regexp.MustCompile("[a-f0-9]{64}")
	deploy := types.DeployArgs{
		Image: "ubuntu",
		GPU: &types.GPU{
			Count: util.Int64Ptr(-1),
		},
	}
	//The logger will not work as the paths are not initialized in the test.
	containerId, err := Run(deploy, types.AuthInfo{})
	if err != nil {
		t.Error(err)
	}
	if !re.MatchString(containerId) {
		t.Errorf("Container not match regex. Got %s, Want format [a-f0-9]{64}", containerId)
	}
	return containerId
}

func inspectTest(containerId string, spec types.DeployArgs, t *testing.T) {
	container, err := Inspect(containerId)
	if err != nil {
		t.Error(err)
	}
	if container.ID != containerId {
		t.Errorf("Container ID incorrect. Got %s, Want %s", container.ID, containerId)
	}
	if container.Image != spec.Image {
		t.Errorf("Image incorrect. Got %s, Want %s", container.Image, spec.Image)
	}
	//A PATH variable is added to the container as it starts. Hence there are 2 env variables
	if len(container.Environments) != 2 {
		t.Errorf("Env var size incorrect. Got %d, Want 2", len(container.Environments))
	} else {
		if container.Environments[0].Name != "test" {
			t.Errorf("Env var key incorrect. Got %s, Want test", container.Environments[0].Name)
		}
		if container.Environments[0].Value != "this is a test" {
			t.Errorf("Env var key incorrect. Got %s, Want this is a test", container.Environments[0].Value)
		}
		if container.ExposePorts[0].Protocol != types.TCP {
			t.Errorf("Expose protocol incorrect. Got %s, Want %s", container.ExposePorts[0].Protocol, types.TCP)
		}
	}
	if container.MemLimit != 512000000 {
		t.Errorf("Memory limit incorrect. Got %d, Want 512000000", container.MemLimit)
	}
	if container.MemSoftLimit != 256000000 {
		t.Errorf("Memory soft limit incorrect. Got %d. Want 256000000", container.MemSoftLimit)
	}
	if container.RestartPolicy != types.RestartOnFailure {
		t.Errorf("Restart policy incorrect. Got %s, Want %s", container.RestartPolicy, types.RestartOnFailure)
	}
	if len(container.Entrypoint) != 1 {
		t.Errorf("Entrypoint size incorrect. Got %d, Want 1", len(container.Entrypoint))
	} else {
		if container.Entrypoint[0] != "bash" {
			t.Errorf("Entrypoint incorrect. Got %s, Want bash", container.Entrypoint[0])
		}
	}
	if container.Command != "bash" {
		t.Errorf("Command incorrect. Got %s, Want bash", container.Command)
	}
	if len(container.ExposePorts) != 1 {
		t.Errorf("Expose port size incorrect. Got %d, Want 1", len(container.ExposePorts))
	} else {
		if container.ExposePorts[0].HostPort != 1234 {
			t.Errorf("Expose host port incorrect. Got %d, Want 1234", container.ExposePorts[0].HostPort)
		}
		if container.ExposePorts[0].ContainerPort != 1234 {
			t.Errorf("Expose container port incorrect. Got %d, Want 1234", container.ExposePorts[0].ContainerPort)
		}
	}
	if len(container.Devices) != 1 {
		t.Errorf("Device size incorrect. Got %d, Want 1", len(container.Devices))
	} else {
		if container.Devices[0].HostDevicePath != "/dev/loop0" {
			t.Errorf("Device host path incorrect. Got %s, Want /dev/loop0", container.Devices[0].HostDevicePath)
		}
		if container.Devices[0].ContainerDevicePath != "/dev/loop0" {
			t.Errorf("Device container path incorrect. Got %s, Want /dev/loop0", container.Devices[0].ContainerDevicePath)
		}
		if container.Devices[0].CgroupPermissions != "r" {
			t.Errorf("Device permissions incorrect. Got %s, Want r", container.Devices[0].CgroupPermissions)
		}
	}
}

func updateTest(containerId string, spec types.DeployArgs, t *testing.T) string {
	spec.ExposePorts[0].ContainerPort = 2345
	spec.ExposePorts[0].HostPort = 2345
	spec.ExposePorts[0].Protocol = types.UDP
	newContainerId, err := Update(containerId, spec, types.AuthInfo{})
	if err != nil {
		t.Error(err)
		return ""
	}
	inspect, _ := Inspect(newContainerId)
	if len(inspect.ExposePorts) != 1 {
		t.Errorf("Expose port size incorrect. Got %d, Want 1", len(inspect.ExposePorts))
	} else {
		if inspect.ExposePorts[0].HostPort != 2345 {
			t.Errorf("Expose host port incorrect. Got %d, Want 2345", inspect.ExposePorts[0].HostPort)
		}
		if inspect.ExposePorts[0].ContainerPort != 2345 {
			t.Errorf("Expose container port incorrect. Got %d, Want 2345", inspect.ExposePorts[0].ContainerPort)
		}
		if inspect.ExposePorts[0].Protocol != types.UDP {
			t.Errorf("Expose protocol incorrect. Got %s, Want %s", inspect.ExposePorts[0].Protocol, types.UDP)
		}
	}
	return newContainerId
}

func stopTest(containers []string, t *testing.T) {
	for _, container := range containers {
		err := Stop(container)
		if err != nil {
			//Ignore if the container is just not running.
			if !strings.Contains(err.Error(), "not running") {
				t.Error(err)
			}
		}
	}
}

func deleteTest(containers []string, t *testing.T) {
	for _, container := range containers {
		err := Delete(container, false)
		if err != nil {
			t.Error(err)
		}
	}
}

func listTest(containerId string, spec types.DeployArgs, t *testing.T) {
	containers, err := ListDetailed()
	if err != nil {
		t.Error("Failed to list containers")
	}
	exist := false
	for _, container := range containers {
		if container.ID == containerId {
			if container.Image != spec.Image {
				t.Errorf("Image incorrect. Got %s, Want %s", container.Image, spec.Image)
			}
			//A PATH variable is added to the container as it starts. Hence there are 2 env variables
			if len(container.Environments) != 2 {
				t.Errorf("Env var size incorrect. Got %d, Want 2", len(container.Environments))
			} else {
				if container.Environments[0].Name != "test" {
					t.Errorf("Env var key incorrect. Got %s, Want test", container.Environments[0].Name)
				}
				if container.Environments[0].Value != "this is a test" {
					t.Errorf("Env var key incorrect. Got %s, Want this is a test", container.Environments[0].Value)
				}
			}
			if container.MemLimit != 512000000 {
				t.Errorf("Memory limit incorrect. Got %d, Want 512000000", container.MemLimit)
			}
			if container.MemSoftLimit != 256000000 {
				t.Errorf("Memory soft limit incorrect. Got %d. Want 256000000", container.MemSoftLimit)
			}
			if container.RestartPolicy != types.RestartOnFailure {
				t.Errorf("Restart policy incorrect. Got %s, Want %s", container.RestartPolicy, types.RestartOnFailure)
			}
			if len(container.Entrypoint) != 1 {
				t.Errorf("Entrypoint size incorrect. Got %d, Want 1", len(container.Entrypoint))
			} else {
				if container.Entrypoint[0] != "bash" {
					t.Errorf("Entrypoint incorrect. Got %s, Want bash", container.Entrypoint[0])
				}
			}
			if container.Command != "bash" {
				t.Errorf("Command incorrect. Got %s, Want bash", container.Command)
			}
			if len(container.ExposePorts) != 1 {
				t.Errorf("Expose port size incorrect. Got %d, Want 1", len(container.ExposePorts))
			} else {
				if container.ExposePorts[0].HostPort != 1234 {
					t.Errorf("Expose host port incorrect. Got %d, Want 1234", container.ExposePorts[0].HostPort)
				}
				if container.ExposePorts[0].ContainerPort != 1234 {
					t.Errorf("Expose container port incorrect. Got %d, Want 1234", container.ExposePorts[0].ContainerPort)
				}
			}
			if len(container.Devices) != 1 {
				t.Errorf("Device size incorrect. Got %d, Want 1", len(container.Devices))
			} else {
				if container.Devices[0].HostDevicePath != "/dev/loop0" {
					t.Errorf("Device host path incorrect. Got %s, Want /dev/loop0", container.Devices[0].HostDevicePath)
				}
				if container.Devices[0].ContainerDevicePath != "/dev/loop0" {
					t.Errorf("Device container path incorrect. Got %s, Want /dev/loop0", container.Devices[0].ContainerDevicePath)
				}
				if container.Devices[0].CgroupPermissions != "r" {
					t.Errorf("Device permissions incorrect. Got %s, Want r", container.Devices[0].CgroupPermissions)
				}
			}
			exist = true
			break
		}
	}
	if !exist {
		t.Error("Container does not exist in list")
	}
}
