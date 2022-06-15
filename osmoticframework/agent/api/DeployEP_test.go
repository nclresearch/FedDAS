package api

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"osmoticframework/agent/docker"
	"osmoticframework/agent/types"
	"regexp"
	"testing"
	"time"
)

var currentContainerId string

func TestEndpoint(t *testing.T) {
	err := docker.Init()
	if err != nil {
		t.Error("Docker failed to initialize")
		t.FailNow()
	}
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
	t.Run("Deploy endpoint", func(t *testing.T) {
		runRequestTest(deploy, t)
	})
	if t.Failed() {
		t.FailNow()
	}
	t.Run("Inspect endpoint", func(t *testing.T) {
		inspectRequestTest(currentContainerId, deploy, t)
	})
	time.Sleep(time.Second)
	if !t.Failed() {
		t.Run("List endpoint", func(t *testing.T) {
			listRequestTest(currentContainerId, t)
		})
		t.Run("Update endpoint", func(t *testing.T) {
			updateRequestTest(currentContainerId, deploy, t)
		})
	} else {
		t.Log("[SKIP] Inspect request test failed. Skipping update request test")
	}
	containers := []string{currentContainerId}
	// Padding time to allow the container to run first
	time.Sleep(time.Second * 5)
	t.Run("Stop endpoint", func(t *testing.T) {
		stopRequestTest(containers, t)
	})
	t.Run("Delete endpoint", func(t *testing.T) {
		deleteRequestTest(containers, t)
	})
}

func runRequestTest(spec types.DeployArgs, t *testing.T) {
	re := regexp.MustCompile("[a-f0-9]{64}")
	var args = map[string]interface{}{}
	args["deployArgs"] = spec
	args["authInfo"] = types.AuthInfo{}
	const expectId = "test"
	const expectStatus = "ok"
	const expectApi = "deploy"
	responseRaw := RunEP(expectId, args)
	var response map[string]string
	err := json.Unmarshal(responseRaw, &response)
	if err != nil {
		t.Error("Failed to unmarshal response")
		return
	}
	if response["requestId"] != expectId {
		t.Errorf("Request ID incorrect. Got %s, Want %s", response["requestId"], expectId)
	}
	if response["status"] != expectStatus {
		t.Errorf("Status incorrect. Got %s, Want %s", response["status"], expectStatus)
	}
	if !re.MatchString(response["containerId"]) {
		t.Errorf("Container ID not matching regex. Got %s, Want [a-f0-9]{64}", response["containerId"])
	}
	if response["api"] != expectApi {
		t.Errorf("API incorrect. Got %s, Want %s", response["api"], expectApi)
	}
	currentContainerId = response["containerId"]
}

func inspectRequestTest(containerId string, spec types.DeployArgs, t *testing.T) {
	var args = map[string]interface{}{}
	args["containerId"] = containerId
	responseRaw := InspectEP("test", args)
	var response map[string]interface{}
	err := json.Unmarshal(responseRaw, &response)
	if err != nil {
		t.Error("Failed to unmarshal response")
		return
	}
	const expectId = "test"
	const expectStatus = "ok"
	const expectApi = "deploy"
	if response["requestId"] != expectId {
		t.Errorf("Request ID incorrect. Got %s, Want %s", response["requestId"], expectId)
	}
	if response["status"] != expectStatus {
		t.Errorf("Status incorrect. Got %s, Want %s", response["status"], expectStatus)
	}
	if response["api"] != expectApi {
		t.Errorf("API incorrect. Got %s, Want %s", response["api"], expectApi)
	}
	var container types.Container
	err = mapstructure.Decode(response["container"], &container)
	if err != nil {
		t.Error("Failed to decode response")
		return
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
		if container.ExposePorts[0].Protocol != types.TCP {
			t.Errorf("Expose protocol incorrect Got %s, Want %s", container.ExposePorts[0].Protocol, types.TCP)
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

func updateRequestTest(containerId string, spec types.DeployArgs, t *testing.T) {
	spec.ExposePorts[0].ContainerPort = 2345
	spec.ExposePorts[0].HostPort = 2345
	spec.ExposePorts[0].Protocol = types.UDP
	var args = map[string]interface{}{}
	args["containerId"] = containerId
	args["deployArgs"] = spec
	args["authInfo"] = types.AuthInfo{}
	responseRaw := UpdateEP("test", args)
	var response map[string]string
	err := json.Unmarshal(responseRaw, &response)
	if err != nil {
		t.Error("Failed to unmarshal response")
		return
	}
	const expectId = "test"
	const expectStatus = "ok"
	const expectApi = "deploy"
	if response["requestId"] != expectId {
		t.Errorf("Request ID incorrect. Got %s, Want %s", response["requestId"], expectId)
	}
	if response["status"] != expectStatus {
		t.Errorf("Status incorrect. Got %s, Want %s", response["status"], expectStatus)
	}
	if response["api"] != expectApi {
		t.Errorf("API incorrect. Got %s, Want %s", response["api"], expectApi)
	}
	newContainerId := response["containerId"]
	inspect, _ := docker.Inspect(newContainerId)
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
	currentContainerId = newContainerId
}

func stopRequestTest(containers []string, t *testing.T) {
	for _, containerId := range containers {
		var args = map[string]interface{}{}
		args["containerId"] = containerId
		responseRaw := StopEP("test", args)
		var response map[string]string
		err := json.Unmarshal(responseRaw, &response)
		if err != nil {
			t.Error("Failed to unmarshal response")
			return
		}
		const expectId = "test"
		const expectStatus = "ok"
		const expectApi = "deploy"
		if response["requestId"] != expectId {
			t.Errorf("Request ID incorrect. Got %s, Want %s", response["requestId"], expectId)
		}
		if response["status"] != expectStatus {
			t.Errorf("Status incorrect. Got %s, Want %s", response["status"], expectStatus)
			t.Error(response["error"])
		}
		if response["api"] != expectApi {
			t.Errorf("API incorrect. Got %s, Want %s", response["api"], expectApi)
		}
	}
}

func deleteRequestTest(containers []string, t *testing.T) {
	for _, containerId := range containers {
		var args = map[string]interface{}{}
		args["containerId"] = containerId
		args["deleteImage"] = false
		responseRaw := DeleteEP("test", args)
		var response map[string]string
		err := json.Unmarshal(responseRaw, &response)
		if err != nil {
			t.Error("Failed to unmarshal response")
			return
		}
		const expectId = "test"
		const expectStatus = "ok"
		const expectApi = "deploy"
		if response["requestId"] != expectId {
			t.Errorf("Request ID incorrect. Got %s, Want %s", response["requestId"], expectId)
		}
		if response["status"] != expectStatus {
			t.Errorf("Status incorrect. Got %s, Want %s", response["status"], expectStatus)
			t.Error(response["error"])
		}
		if response["api"] != expectApi {
			t.Errorf("API incorrect. Got %s, Want %s", response["api"], expectApi)
		}
	}
}

func listRequestTest(containerId string, t *testing.T) {
	responseRaw := ListEP("test")
	var response map[string]interface{}
	err := json.Unmarshal(responseRaw, &response)
	if err != nil {
		t.Error("Failed to unmarshal response")
		return
	}
	const expectId = "test"
	const expectStatus = "ok"
	const expectApi = "deploy"
	if response["requestId"] != expectId {
		t.Errorf("Request ID incorrect. Got %s, Want %s", response["requestId"], expectId)
	}
	if response["status"] != expectStatus {
		t.Errorf("Status incorrect. Got %s, Want %s", response["status"], expectStatus)
	}
	if response["api"] != expectApi {
		t.Errorf("API incorrect. Got %s, Want %s", response["api"], expectApi)
	}
	var containers []types.Container
	err = mapstructure.Decode(response["containers"], &containers)
	if err != nil {
		t.Error("Failed to decode response")
	}
	//As tests are running in parallel, we cannot accurately test how many containers are deployed
	//We'll just check if the deployed container exists in the API response
	exist := false
	for _, container := range containers {
		if container.ID == containerId {
			exist = true
			break
		}
	}
	if !exist {
		t.Error("Container does not exist in list request")
	}
}
