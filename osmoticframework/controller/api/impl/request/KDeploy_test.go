package request

import (
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"os"
	"osmoticframework/controller/types"
	"osmoticframework/controller/vars"
	"reflect"
	"testing"
	"time"
)

func TestEndpoints(t *testing.T) {
	home, _ := os.UserHomeDir()
	//Change the config path as needed
	var config = `
{
  "rabbitAddress": "",
  "databaseAddress": "",
  "prometheusAddress": "",
  "kuberConfigPath": "` + home + `/.kube/config",
  "networks": []
}
`
	vars.LoadConfig([]byte(config))
	client := getKuber()
	if client == nil {
		t.Error("Failed to get client. client is nil")
		t.FailNow()
	}
	deployName := "ubuntu"
	deploy := types.KDeployArgs{
		PodArgs: types.KPodArgs{
			Image: "ubuntu",
			Label: map[string]string{
				"app": deployName,
			},
			Name: deployName,
			ExposePorts: []types.ExposePort{
				{
					HostPort:      1234,
					ContainerPort: 1234,
					Protocol:      types.TCP,
				},
			},
			Entrypoint: []string{"bash"},
			Arguments:  []string{"bash"},
			Environment: []types.Environment{
				{
					Name:  "test",
					Value: "this is a test",
				},
			},
			MemLimit:     1000000000,
			MemSoftLimit: 512000000,
			CPULimit:     2,
			CPUSoftLimit: 1,
			Volumes: []types.KVolume{
				{
					MountPoint: "/rootfs",
					ReadOnly:   true,
					Type:       types.TypeHostPath,
					HostPath: &types.HostPath{
						Path:         "/",
						MountType:    corev1.HostPathDirectory,
						NodeHostname: "cloud-server",
					},
				},
				{
					MountPoint: "/aws",
					ReadOnly:   true,
					Type:       types.TypeAWSEBS,
					AwsEbs: &types.AwsEbs{
						VolumeID:  "Test",
						FSType:    "ext4",
						Partition: 1,
					},
				},
			},
			PullPolicy: corev1.PullIfNotPresent,
		},
		DeploymentName: deployName,
		Replicas:       1,
		Strategy:       v1.RecreateDeploymentStrategyType,
	}
	jobName := "fedora"
	job := types.KJobArgs{
		PodArgs: types.KPodArgs{
			Image: "fedora",
			Label: map[string]string{
				"app": jobName,
			},
			Name: jobName,
			ExposePorts: []types.ExposePort{
				{
					HostPort:      3456,
					ContainerPort: 3456,
					Protocol:      types.TCP,
				},
			},
			Entrypoint: []string{"bash"},
			Arguments:  []string{"-c", "sleep 60s"},
			Environment: []types.Environment{
				{
					Name:  "test",
					Value: "this is a test",
				},
			},
			MemLimit:     1000000000,
			MemSoftLimit: 512000000,
			CPULimit:     2,
			CPUSoftLimit: 1,
			Volumes: []types.KVolume{
				{
					MountPoint: "/opt/nfs/registry-storage",
					ReadOnly:   true,
					Type:       types.TypeNFS,
					NFS: &types.NFS{
						Server: "localhost",
						Path:   "/serve",
					},
				},
				{
					MountPoint: "/aws",
					ReadOnly:   true,
					Type:       types.TypeAWSEBS,
					AwsEbs: &types.AwsEbs{
						VolumeID:  "Test",
						FSType:    "ext4",
						Partition: 1,
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
			PullPolicy:    corev1.PullIfNotPresent,
		},
		JobName:      jobName,
		BackOffLimit: 4,
	}
	serviceName := "ubuntu-service"
	service := types.KServiceArgs{
		Name: serviceName,
		Ports: []types.KServicePort{
			{
				Name:       "test-port",
				Host:       1234,
				TargetPort: 1234,
				NodePort:   31234,
				Protocol:   types.TCP,
			},
		},
		Selector: map[string]string{},
		Type:     corev1.ServiceTypeNodePort,
	}
	cronjobName := "debian"
	cronjob := types.KCronjobArgs{
		JobArgs: types.KJobArgs{
			PodArgs: types.KPodArgs{
				Image: "debian",
				Label: map[string]string{
					"app": cronjobName,
				},
				Name: cronjobName,
				ExposePorts: []types.ExposePort{
					{
						HostPort:      4567,
						ContainerPort: 4567,
						Protocol:      types.TCP,
					},
				},
				Entrypoint: []string{"bash"},
				Arguments:  []string{"-c", "sleep 60s"},
				Environment: []types.Environment{
					{
						Name:  "test",
						Value: "this is a test",
					},
				},
				MemLimit:     1000000000,
				MemSoftLimit: 512000000,
				CPULimit:     2,
				CPUSoftLimit: 1,
				Volumes: []types.KVolume{
					{
						MountPoint: "/opt/nfs/registry-storage",
						ReadOnly:   true,
						Type:       types.TypeNFS,
						NFS: &types.NFS{
							Server: "localhost",
							Path:   "/serve",
						},
					},
					{
						MountPoint: "/aws",
						ReadOnly:   true,
						Type:       types.TypeAWSEBS,
						AwsEbs: &types.AwsEbs{
							VolumeID:  "Test",
							FSType:    "ext4",
							Partition: 1,
						},
					},
				},
				PullPolicy:    corev1.PullIfNotPresent,
				RestartPolicy: corev1.RestartPolicyOnFailure,
			},
			JobName:      jobName,
			BackOffLimit: 4,
		},
		JobName: cronjobName,
		Cron:    "*/10 * * * *",
	}
	const configMapName = "test-config"
	configMap := types.KConfigMap{
		Name: configMapName,
		Data: map[string]string{
			"test.conf": "HELLO=true",
		},
	}
	t.Run("Run endpoint", func(t *testing.T) {
		runTest(deploy, t)
	})
	if t.Failed() {
		t.FailNow()
	}
	t.Run("Get deployment endpoint", func(t *testing.T) {
		getDeploymentTest(deploy, t)
	})
	t.Run("List deployment endpoint", func(t *testing.T) {
		listDeploymentTest(deploy, t)
	})
	failed := t.Failed()
	t.Run("Update deployment endpoint", func(t *testing.T) {
		if failed {
			t.Skip("Skip due to previous tests failed")
		}
		deploy = updateDeploymentTest(deploy, t)
	})
	t.Run("Delete endpoint", func(t *testing.T) {
		deleteTest(deployName, t)
	})
	t.Run("Create service endpoint", func(t *testing.T) {
		service = createServiceTest(service, t)
	})
	t.Run("Get service endpoint", func(t *testing.T) {
		getServiceTest(service, t)
	})
	t.Run("List service endpoint", func(t *testing.T) {
		listServiceTest(service, t)
	})
	failed = t.Failed()
	t.Run("Update service endpoint", func(t *testing.T) {
		if failed {
			t.Skip("Skip due to previous tests failed")
		}
		service = updateServiceTest(service, t)
	})
	t.Run("Delete service endpoint", func(t *testing.T) {
		deleteServiceTest(serviceName, t)
	})
	t.Run("Run job endpoint", func(t *testing.T) {
		runJobTest(job, t)
	})
	if t.Failed() {
		t.FailNow()
	}
	t.Run("Get job endpoint", func(t *testing.T) {
		getJobTest(job, t)
	})
	t.Run("List job endpoint", func(t *testing.T) {
		listJobTest(job, t)
	})
	t.Run("Delete job endpoint", func(t *testing.T) {
		deleteJobTest(jobName, t)
	})
	t.Run("Run cronjob endpoint", func(t *testing.T) {
		runCronjobTest(cronjob, t)
	})
	if t.Failed() {
		t.FailNow()
	}
	t.Run("Get cronjob endpoint", func(t *testing.T) {
		getCronjobTest(cronjob, t)
	})
	t.Run("List cronjob endpoint", func(t *testing.T) {
		listCronjobTest(cronjob, t)
	})
	t.Run("Delete cronjob endpoint", func(t *testing.T) {
		deleteCronjobTest(cronjobName, t)
	})
	t.Run("Create config map", func(t *testing.T) {
		createConfigMapTest(configMap, t)
	})
	t.Run("List config map", func(t *testing.T) {
		listConfigMapTest(t)
	})
	t.Run("Delete config map", func(t *testing.T) {
		deleteConfigMapTest(configMapName, t)
	})
}

func runTest(spec types.KDeployArgs, t *testing.T) {
	err := KRunDeployment(spec, []string{})
	if err != nil {
		t.Errorf("Failed sending run request, %s", err)
	}
}

func getDeploymentTest(deploy types.KDeployArgs, t *testing.T) {
	spec, err := KGetDeployment(deploy.DeploymentName)
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
	if spec.PodArgs.Image != deploy.PodArgs.Image {
		t.Errorf("Image incorrect. Got %s, Want %s", spec.PodArgs.Image, deploy.PodArgs.Image)
	}
	if spec.PodArgs.Name != deploy.PodArgs.Name {
		t.Errorf("Name incorrect. Got %s, Want %s", spec.PodArgs.Name, deploy.PodArgs.Name)
	}
	if spec.DeploymentName != deploy.DeploymentName {
		t.Errorf("Deployment name incorrect. Got %s, Want %s", spec.DeploymentName, deploy.DeploymentName)
	}
	if spec.PodArgs.MemLimit != deploy.PodArgs.MemLimit {
		t.Errorf("Memory limit incorrect. Got %d, Want %d", spec.PodArgs.MemLimit, deploy.PodArgs.MemLimit)
	}
	if spec.PodArgs.MemSoftLimit != deploy.PodArgs.MemSoftLimit {
		t.Errorf("Memory soft limit incorrect. Got %d, Want %d", spec.PodArgs.MemSoftLimit, deploy.PodArgs.MemSoftLimit)
	}
	if spec.PodArgs.PullPolicy != deploy.PodArgs.PullPolicy {
		t.Errorf("Pull policy incorrect. Got %s, Want %s", spec.PodArgs.PullPolicy, deploy.PodArgs.PullPolicy)
	}
	if spec.PodArgs.CPULimit != deploy.PodArgs.CPULimit {
		t.Errorf("CPU limit incorrect. Got %f, Want %f", spec.PodArgs.CPULimit, deploy.PodArgs.CPULimit)
	}
	if spec.PodArgs.CPUSoftLimit != deploy.PodArgs.CPUSoftLimit {
		t.Errorf("CPU soft limit incorrect. Got %f, Want %f", spec.PodArgs.CPUSoftLimit, deploy.PodArgs.CPUSoftLimit)
	}
	if spec.Strategy != deploy.Strategy {
		t.Errorf("Strategy incorrect. Got %s, Want %s", spec.Strategy, deploy.Strategy)
	}
	if spec.Replicas != deploy.Replicas {
		t.Errorf("Replicas incorrect. Got %d, Want %d", spec.Replicas, deploy.Replicas)
	}
	var expectArgs = []string{"bash"}
	if !reflect.DeepEqual(spec.PodArgs.Arguments, expectArgs) {
		t.Errorf("Arguments incorrect. Got %s, Want %s", spec.PodArgs.Arguments, expectArgs)
	}
	var expectEntry = []string{"bash"}
	if !reflect.DeepEqual(spec.PodArgs.Entrypoint, expectEntry) {
		t.Errorf("Entrypoint incorrect. Got %s, Want %s", spec.PodArgs.Entrypoint, expectEntry)
	}
	var expectVolume = []types.KVolume{
		{
			MountPoint: "/rootfs",
			ReadOnly:   true,
			Type:       types.TypeHostPath,
			HostPath: &types.HostPath{
				Path:         "/",
				MountType:    corev1.HostPathDirectory,
				NodeHostname: "cloud-server",
			},
		},
		{
			MountPoint: "/aws",
			ReadOnly:   true,
			Type:       types.TypeAWSEBS,
			AwsEbs: &types.AwsEbs{
				VolumeID:  "Test",
				FSType:    "ext4",
				Partition: 1,
			},
		},
	}
	if !reflect.DeepEqual(spec.PodArgs.Volumes, expectVolume) {
		t.Errorf("Volumes incorrect. Got %#v, Want %#v", spec.PodArgs.Volumes, expectVolume)
	}
	var expectPorts = []types.ExposePort{
		{
			HostPort:      1234,
			ContainerPort: 1234,
			Protocol:      types.TCP,
		},
	}
	if !reflect.DeepEqual(spec.PodArgs.ExposePorts, expectPorts) {
		t.Errorf("Exposed ports incorrect. Got %#v, Want %#v", spec.PodArgs.ExposePorts, expectPorts)
	}
	if len(spec.PodArgs.Label) != 2 {
		t.Errorf("Label size incorrect. Got %d, Want 2", len(spec.PodArgs.Label))
	}
	if spec.PodArgs.Label["app"] != deploy.DeploymentName {
		t.Errorf("Label incorrect. Got %s, Want %s", spec.PodArgs.Label["app"], deploy.DeploymentName)
	}
	//Unlike Docker, get deployment does not return the container's current environment variables. (Which includes PATH)
	//So the env var length is 1
	if len(spec.PodArgs.Environment) != 1 {
		t.Errorf("Environment size incorrect. Got %d, Want %d", len(spec.PodArgs.Environment), 1)
		t.Error(spec.PodArgs.Environment)
	} else {
		if spec.PodArgs.Environment[0].Name != "test" {
			t.Errorf("EnvVar name incorrect. Got %s, Want %s", spec.PodArgs.Environment[0].Name, "test")
		}
		if spec.PodArgs.Environment[0].Value != "this is a test" {
			t.Errorf("EnvVar value incorrect. Got %s, Want %s", spec.PodArgs.Environment[0].Value, "this is a test")
		}
	}
}

func listDeploymentTest(deploy types.KDeployArgs, t *testing.T) {
	deployments, err := KListDeployment()
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
	exist := false
	for _, deployment := range deployments {
		if deployment.PodArgs.Name == deploy.PodArgs.Name {
			exist = true
			break
		}
	}
	if !exist {
		t.Error("Deployment does not exist in list")
	}
}

func updateDeploymentTest(deploy types.KDeployArgs, t *testing.T) types.KDeployArgs {
	deploy.PodArgs.ExposePorts[0].ContainerPort = 2345
	deploy.PodArgs.ExposePorts[0].HostPort = 2345
	deploy.PodArgs.ExposePorts[0].Protocol = types.UDP
	err := KUpdateDeployment(deploy, []string{})
	if err != nil {
		t.Errorf("Request failed, %s", err)
	}
	deployment, err := KGetDeployment(deploy.PodArgs.Name)
	if err != nil {
		t.Errorf("Cannot verify update status, %s", err)
		return deploy
	}
	hostPort := deployment.PodArgs.ExposePorts[0].HostPort
	containerPort := deployment.PodArgs.ExposePorts[0].ContainerPort
	protocol := deployment.PodArgs.ExposePorts[0].Protocol
	if hostPort != 2345 {
		t.Errorf("Host port incorrect. Got %d, Want %d", hostPort, 2345)
	}
	if containerPort != 2345 {
		t.Errorf("Container port incorrect. Got %d, Want %d", containerPort, 2345)
	}
	if protocol != types.UDP {
		t.Errorf("Protocol incorrect. Got %s, Want %s", protocol, types.UDP)
	}
	return deploy
}

func createServiceTest(service types.KServiceArgs, t *testing.T) types.KServiceArgs {
	deployedService, err := KCreateService(service)
	if err != nil {
		t.Errorf("Request failed, %s", err)
	}
	return *deployedService
}

func getServiceTest(service types.KServiceArgs, t *testing.T) {
	deployedService, err := KGetService(service.Name)
	if err != nil {
		t.Errorf("Request failed, %s", err)
	}
	if deployedService.Name != service.Name {
		t.Errorf("Name incorrect. Got %s, Want %s", deployedService.Name, service.Name)
	}
	if deployedService.Type != service.Type {
		t.Errorf("Service type incorrect. Got %s, Want %s", deployedService.Type, service.Type)
	}
	if !reflect.DeepEqual(deployedService.Selector, service.Selector) {
		t.Errorf("Service selector incorrect. Got %#v, Want %#v", deployedService.Selector, service.Selector)
	}
	if !reflect.DeepEqual(deployedService.Ports, service.Ports) {
		t.Errorf("Ports incorrect. Got %#v, Want %#v", deployedService.Ports, service.Ports)
	}
}

func listServiceTest(service types.KServiceArgs, t *testing.T) {
	deployedServices, err := KListService()
	if err != nil {
		t.Errorf("Request failed, %s", err)
	}
	exist := false
	for _, deployedService := range deployedServices {
		if deployedService.Name == service.Name {
			exist = true
			break
		}
	}
	if !exist {
		t.Error("Service does not exist in list")
	}
}

func updateServiceTest(service types.KServiceArgs, t *testing.T) types.KServiceArgs {
	service.Ports[0].NodePort = 32345
	service.Ports[0].TargetPort = 2345
	service.Ports[0].Host = 2345
	service.Ports[0].Protocol = types.UDP
	_, err := KUpdateService(service)
	if err != nil {
		t.Errorf("Request failed, %s", err)
	}
	deployedService, err := KGetService(service.Name)
	if err != nil {
		t.Errorf("Failed to verify updated service, %s", err)
	}
	nodePort := deployedService.Ports[0].NodePort
	hostPort := deployedService.Ports[0].Host
	targetPort := deployedService.Ports[0].TargetPort
	protocol := deployedService.Ports[0].Protocol
	if nodePort != 32345 {
		t.Errorf("Node port incorrect. Got %d, Want %d", nodePort, 32345)
	}
	if hostPort != 2345 {
		t.Errorf("Host port incorrect. Got %d, Want %d", hostPort, 2345)
	}
	if targetPort != 2345 {
		t.Errorf("Target port incorrect. Got %d, Want %d", targetPort, 2345)
	}
	if protocol != types.UDP {
		t.Errorf("Protocol incorrect. Got %s, Want %s", protocol, types.UDP)
	}
	return service
}

func deleteTest(deployName string, t *testing.T) {
	err := KDeleteDeployment(deployName)
	if err != nil {
		t.Errorf("Failed sneding delete request, %s", err)
	}
}

func deleteServiceTest(serviceName string, t *testing.T) {
	err := KDeleteService(serviceName)
	if err != nil {
		t.Errorf("Failed sneding delete request, %s", err)
	}
}

func runJobTest(spec types.KJobArgs, t *testing.T) {
	err := KRunJob(spec, []string{})
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
}

func getJobTest(job types.KJobArgs, t *testing.T) {
	t.Log("Wait for 10 seconds for pod to deploy")
	time.Sleep(time.Second * 10)
	spec, err := KGetJob(job.JobName)
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
	if spec.PodArgs.Image != job.PodArgs.Image {
		t.Errorf("Image incorrect. Got %s, Want %s", spec.PodArgs.Image, job.PodArgs.Image)
	}
	if spec.PodArgs.Name != job.PodArgs.Name {
		t.Errorf("Name incorrect. Got %s, Want %s", spec.PodArgs.Name, job.PodArgs.Name)
	}
	if spec.JobName != job.JobName {
		t.Errorf("Deployment name incorrect. Got %s, Want %s", spec.JobName, job.JobName)
	}
	if spec.PodArgs.MemLimit != job.PodArgs.MemLimit {
		t.Errorf("Memory limit incorrect. Got %d, Want %d", spec.PodArgs.MemLimit, job.PodArgs.MemLimit)
	}
	if spec.PodArgs.MemSoftLimit != job.PodArgs.MemSoftLimit {
		t.Errorf("Memory soft limit incorrect. Got %d, Want %d", spec.PodArgs.MemSoftLimit, job.PodArgs.MemSoftLimit)
	}
	if spec.PodArgs.PullPolicy != job.PodArgs.PullPolicy {
		t.Errorf("Pull policy incorrect. Got %s, Want %s", spec.PodArgs.PullPolicy, job.PodArgs.PullPolicy)
	}
	if spec.PodArgs.CPULimit != job.PodArgs.CPULimit {
		t.Errorf("CPU limit incorrect. Got %f, Want %f", spec.PodArgs.CPULimit, job.PodArgs.CPULimit)
	}
	if spec.PodArgs.CPUSoftLimit != job.PodArgs.CPUSoftLimit {
		t.Errorf("CPU soft limit incorrect. Got %f, Want %f", spec.PodArgs.CPUSoftLimit, job.PodArgs.CPUSoftLimit)
	}
	if spec.BackOffLimit != job.BackOffLimit {
		t.Errorf("Back off limit incorrect. Got %d, Want %d", spec.BackOffLimit, job.BackOffLimit)
	}
	var expectArgs = []string{"-c", "sleep 60s"}
	if !reflect.DeepEqual(spec.PodArgs.Arguments, expectArgs) {
		t.Errorf("Arguments incorrect. Got %s, Want %s", spec.PodArgs.Arguments, expectArgs)
	}
	var expectEntry = []string{"bash"}
	if !reflect.DeepEqual(spec.PodArgs.Entrypoint, expectEntry) {
		t.Errorf("Entrypoint incorrect. Got %s, Want %s", spec.PodArgs.Entrypoint, expectEntry)
	}
	var expectVolume = []types.KVolume{
		{
			MountPoint: "/opt/nfs/registry-storage",
			ReadOnly:   true,
			Type:       types.TypeNFS,
			NFS: &types.NFS{
				Server: "localhost",
				Path:   "/serve",
			},
		},
		{
			MountPoint: "/aws",
			ReadOnly:   true,
			Type:       types.TypeAWSEBS,
			AwsEbs: &types.AwsEbs{
				VolumeID:  "Test",
				FSType:    "ext4",
				Partition: 1,
			},
		},
	}
	if !reflect.DeepEqual(spec.PodArgs.Volumes, expectVolume) {
		t.Errorf("Volumes incorrect. Got %#v, Want %#v", spec.PodArgs.Volumes, expectVolume)
	}
	var expectPorts = []types.ExposePort{
		{
			HostPort:      3456,
			ContainerPort: 3456,
			Protocol:      types.TCP,
		},
	}
	if !reflect.DeepEqual(spec.PodArgs.ExposePorts, expectPorts) {
		t.Errorf("Exposed ports incorrect. Got %#v, Want %#v", spec.PodArgs.ExposePorts, expectPorts)
	}
	if len(spec.PodArgs.Label) != 1 {
		t.Errorf("Label size incorrect. Got %d, Want 1", len(spec.PodArgs.Label))
	}
	if spec.PodArgs.Label["app"] != job.JobName {
		t.Errorf("Label incorrect. Got %s, Want %s", spec.PodArgs.Label["app"], job.JobName)
	}
	//Unlike Docker, get deployment does not return the container's current environment variables. (Which includes PATH)
	//So the env var length is 1
	if len(spec.PodArgs.Environment) != 1 {
		t.Errorf("Environment size incorrect. Got %d, Want %d", len(spec.PodArgs.Environment), 1)
		t.Error(spec.PodArgs.Environment)
	} else {
		if spec.PodArgs.Environment[0].Name != "test" {
			t.Errorf("EnvVar name incorrect. Got %s, Want %s", spec.PodArgs.Environment[0].Name, "test")
		}
		if spec.PodArgs.Environment[0].Value != "this is a test" {
			t.Errorf("EnvVar value incorrect. Got %s, Want %s", spec.PodArgs.Environment[0].Value, "this is a test")
		}
	}
	if spec.PodArgs.RestartPolicy != corev1.RestartPolicyNever {
		t.Errorf("Restart policy incorrect. Got %s, Want %s", spec.PodArgs.RestartPolicy, corev1.RestartPolicyNever)
	}
}

func listJobTest(spec types.KJobArgs, t *testing.T) {
	jobs, err := KListJob()
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
	exist := false
	for _, job := range jobs {
		if job.PodArgs.Name == spec.PodArgs.Name {
			exist = true
			break
		}
	}
	if !exist {
		t.Error("Job does not exist in list")
	}
}

func deleteJobTest(jobName string, t *testing.T) {
	err := KDeleteJob(jobName)
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
}

func runCronjobTest(spec types.KCronjobArgs, t *testing.T) {
	err := KRunCronjob(spec, []string{})
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
}

func getCronjobTest(cronjob types.KCronjobArgs, t *testing.T) {
	t.Log("Wait for 10 seconds for pod to deploy")
	time.Sleep(time.Second * 10)
	spec, err := KGetCronjob(cronjob.JobName)
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
	if spec.Cron != cronjob.Cron {
		t.Errorf("Cron schedule incorrect. Got %s, Want %s", spec.Cron, cronjob.Cron)
	}
	if spec.JobArgs.PodArgs.Image != cronjob.JobArgs.PodArgs.Image {
		t.Errorf("Image incorrect. Got %s, Want %s", spec.JobArgs.PodArgs.Image, cronjob.JobArgs.PodArgs.Image)
	}
	if spec.JobArgs.PodArgs.Name != cronjob.JobArgs.PodArgs.Name {
		t.Errorf("Name incorrect. Got %s, Want %s", spec.JobArgs.PodArgs.Name, cronjob.JobArgs.PodArgs.Name)
	}
	if spec.JobName != cronjob.JobName {
		t.Errorf("Deployment name incorrect. Got %s, Want %s", spec.JobName, cronjob.JobName)
	}
	if spec.JobArgs.PodArgs.MemLimit != cronjob.JobArgs.PodArgs.MemLimit {
		t.Errorf("Memory limit incorrect. Got %d, Want %d", spec.JobArgs.PodArgs.MemLimit, cronjob.JobArgs.PodArgs.MemLimit)
	}
	if spec.JobArgs.PodArgs.MemSoftLimit != cronjob.JobArgs.PodArgs.MemSoftLimit {
		t.Errorf("Memory soft limit incorrect. Got %d, Want %d", spec.JobArgs.PodArgs.MemSoftLimit, cronjob.JobArgs.PodArgs.MemSoftLimit)
	}
	if spec.JobArgs.PodArgs.PullPolicy != cronjob.JobArgs.PodArgs.PullPolicy {
		t.Errorf("Pull policy incorrect. Got %s, Want %s", spec.JobArgs.PodArgs.PullPolicy, cronjob.JobArgs.PodArgs.PullPolicy)
	}
	if spec.JobArgs.PodArgs.CPULimit != cronjob.JobArgs.PodArgs.CPULimit {
		t.Errorf("CPU limit incorrect. Got %f, Want %f", spec.JobArgs.PodArgs.CPULimit, cronjob.JobArgs.PodArgs.CPULimit)
	}
	if spec.JobArgs.PodArgs.CPUSoftLimit != cronjob.JobArgs.PodArgs.CPUSoftLimit {
		t.Errorf("CPU soft limit incorrect. Got %f, Want %f", spec.JobArgs.PodArgs.CPUSoftLimit, cronjob.JobArgs.PodArgs.CPUSoftLimit)
	}
	if spec.JobArgs.BackOffLimit != cronjob.JobArgs.BackOffLimit {
		t.Errorf("Back off limit incorrect. Got %d, Want %d", spec.JobArgs.BackOffLimit, cronjob.JobArgs.BackOffLimit)
	}
	var expectArgs = []string{"-c", "sleep 60s"}
	if !reflect.DeepEqual(spec.JobArgs.PodArgs.Arguments, expectArgs) {
		t.Errorf("Arguments incorrect. Got %s, Want %s", spec.JobArgs.PodArgs.Arguments, expectArgs)
	}
	var expectEntry = []string{"bash"}
	if !reflect.DeepEqual(spec.JobArgs.PodArgs.Entrypoint, expectEntry) {
		t.Errorf("Entrypoint incorrect. Got %s, Want %s", spec.JobArgs.PodArgs.Entrypoint, expectEntry)
	}
	var expectVolume = []types.KVolume{
		{
			MountPoint: "/opt/nfs/registry-storage",
			ReadOnly:   true,
			Type:       types.TypeNFS,
			NFS: &types.NFS{
				Server: "localhost",
				Path:   "/serve",
			},
		},
		{
			MountPoint: "/aws",
			ReadOnly:   true,
			Type:       types.TypeAWSEBS,
			AwsEbs: &types.AwsEbs{
				VolumeID:  "Test",
				FSType:    "ext4",
				Partition: 1,
			},
		},
	}
	if !reflect.DeepEqual(spec.JobArgs.PodArgs.Volumes, expectVolume) {
		t.Errorf("Volumes incorrect. Got %#v, Want %#v", spec.JobArgs.PodArgs.Volumes, expectVolume)
	}
	var expectPorts = []types.ExposePort{
		{
			HostPort:      4567,
			ContainerPort: 4567,
			Protocol:      types.TCP,
		},
	}
	if !reflect.DeepEqual(spec.JobArgs.PodArgs.ExposePorts, expectPorts) {
		t.Errorf("Exposed ports incorrect. Got %#v, Want %#v", spec.JobArgs.PodArgs.ExposePorts, expectPorts)
	}
	if spec.JobArgs.PodArgs.Label["app"] != cronjob.JobName {
		t.Errorf("Label incorrect. Got %s, Want %s", spec.JobArgs.PodArgs.Label["app"], cronjob.JobName)
		t.Errorf("Label contents: %#v", spec.JobArgs.PodArgs.Label)
	}
	//Unlike Docker, get deployment does not return the container's current environment variables. (Which includes PATH)
	//So the env var length is 1
	if len(spec.JobArgs.PodArgs.Environment) != 1 {
		t.Errorf("Environment size incorrect. Got %d, Want %d", len(spec.JobArgs.PodArgs.Environment), 1)
		t.Error(spec.JobArgs.PodArgs.Environment)
	} else {
		if spec.JobArgs.PodArgs.Environment[0].Name != "test" {
			t.Errorf("EnvVar name incorrect. Got %s, Want %s", spec.JobArgs.PodArgs.Environment[0].Name, "test")
		}
		if spec.JobArgs.PodArgs.Environment[0].Value != "this is a test" {
			t.Errorf("EnvVar value incorrect. Got %s, Want %s", spec.JobArgs.PodArgs.Environment[0].Value, "this is a test")
		}
	}
	if spec.JobArgs.PodArgs.RestartPolicy != corev1.RestartPolicyOnFailure {
		t.Errorf("Restart policy incorrect. Got %s, Want %s", spec.JobArgs.PodArgs.RestartPolicy, corev1.RestartPolicyOnFailure)
	}
}

func listCronjobTest(spec types.KCronjobArgs, t *testing.T) {
	cronjobs, err := KListCronjob()
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
	exist := false
	for _, cronjob := range cronjobs {
		if cronjob.JobArgs.PodArgs.Name == spec.JobArgs.PodArgs.Name {
			exist = true
			break
		}
	}
	if !exist {
		t.Error("Job does not exist in list")
	}
}

func deleteCronjobTest(cronjobName string, t *testing.T) {
	err := KDeleteCronJob(cronjobName)
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
}

func createConfigMapTest(config types.KConfigMap, t *testing.T) {
	err := KCreateConfigMap(config)
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
}

func listConfigMapTest(t *testing.T) {
	configMaps, err := KListConfigMap()
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
	if len(configMaps) == 0 {
		t.Error("ConfigMap list is empty")
	}
	for _, configMap := range configMaps {
		if configMap.Name == "test-config" {
			return
		}
	}
	t.Error("ConfigMap does not exist in list")
}

func deleteConfigMapTest(configMapName string, t *testing.T) {
	err := KDeleteConfigMap(configMapName)
	if err != nil {
		t.Errorf("Failed sending request, %s", err)
	}
}
