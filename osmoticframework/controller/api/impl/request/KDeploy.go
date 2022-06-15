package request

import (
	"context"
	"errors"
	"fmt"
	"github.com/lithammer/shortuuid"
	"k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1b "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"osmoticframework/controller/log"
	"osmoticframework/controller/types"
	"osmoticframework/controller/util"
	"osmoticframework/controller/vars"
	"strings"
	"sync"
)

/*
All of the deploy API request functions.
For pulling images using secrets, you'll need to use kubectl to create the secret. Then specify the name of the secret in the argument.
see https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry
*/

//Starts a deployment in Kubernetes
//Kubernetes will automatically pull all of the images and dependencies required
func KRunDeployment(deployArgs types.KDeployArgs, secrets []string) error {
	kuber := getKuber()
	//Replicas cannot be 0. Otherwise Kubernetes will create a deployment with no pods. Wasting resources
	if deployArgs.Replicas <= 0 {
		return errors.New("replicas cannot be lower than 1")
	}
	//There are 3 labels in a deployment
	//1 for the deployment itself (metadata.labels)
	//1 for the pod (spec.template.spec.labels)
	//1 to determine how the deployment will match with the pod, so that resources can be allocated (spec.selector.matchLabels.label)
	label := make(map[string]string)
	for k, v := range deployArgs.PodArgs.Label {
		label[k] = v
	}
	//We add a UUID to the label for the first two labels so that we can monitor using these labels as selectors
	label["deploymentId"] = shortuuid.New()
	podSpec, err := buildPodSpec(deployArgs.DeploymentName, deployArgs.PodArgs, secrets)
	if err != nil {
		return err
	}
	deployment := v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "app/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   deployArgs.DeploymentName,
			Labels: label,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &deployArgs.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: deployArgs.PodArgs.Label},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: label},
				Spec:       *podSpec,
			},
			Strategy:             v1.DeploymentStrategy{Type: deployArgs.Strategy},
			RevisionHistoryLimit: util.Int32Ptr(0),
		},
		Status: v1.DeploymentStatus{},
	}
	//Deployments cannot be stopped. This is the only valid value
	deployment.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyAlways
	//Let Kubernetes handle any other invalid entries
	_, err = kuber.AppsV1().Deployments("default").Create(context.Background(), &deployment, metav1.CreateOptions{})
	//The Go library for Kubernetes is based on REST
	//The function returns immediately after the container is made but Kubernetes doesn't care if the container starts correctly
	//An error only occurs at the REST level (Input format errors, connection failure, etc.)
	//It does not return an error if the container configurations are incorrect (Invalid image, container crashes, etc.)
	//To check if the container is deployed correctly. You must inspect the deployment

	//If we need to know all currently deployed deployments, we can simply ask Kubernetes instead of a database

	if err != nil {
		log.Error.Println("Create deployment " + deployArgs.DeploymentName + " failed")
		log.Error.Println(err)
	}
	log.Info.Printf("Deployment %s created\n", deployArgs.DeploymentName)
	vars.Deployments.Store(deployArgs.DeploymentName, nil)
	return err
}

//Deletes a deployment in Kubernetes
//Kubernetes has a garbage collector to remove unused images. So we don't need to manually remove them.
func KDeleteDeployment(deployName string) error {
	kuber := getKuber()
	err := kuber.AppsV1().Deployments("default").Delete(context.Background(), deployName, metav1.DeleteOptions{})

	if err != nil {
		log.Error.Println("Delete deployment " + deployName + " failed")
		log.Error.Println(err)
	}
	log.Info.Printf("Deployment %s deleted\n", deployName)
	vars.Deployments.Delete(deployName)
	return err
}

//Updates a deployment in Kubernetes
//Note: The deployment name must match existing deployments. Otherwise this will not work.
func KUpdateDeployment(deployArgs types.KDeployArgs, secrets []string) error {
	kuber := getKuber()
	//Replicas cannot be 0. Otherwise Kubernetes will create a deployment with no pods. Wasting resources
	if deployArgs.Replicas <= 0 {
		return errors.New("replicas cannot be lower than 1")
	}
	resources := corev1.ResourceRequirements{
		Limits:   make(corev1.ResourceList),
		Requests: make(corev1.ResourceList),
	}
	//Only fill in information if it exists. Otherwise Kubernetes will put 0 for the limits instead of automatic.
	if deployArgs.PodArgs.MemLimit != 0 {
		resources.Limits[corev1.ResourceMemory] = *resource.NewQuantity(deployArgs.PodArgs.MemLimit, resource.DecimalSI)
	}
	if deployArgs.PodArgs.MemSoftLimit != 0 {
		resources.Requests[corev1.ResourceMemory] = *resource.NewQuantity(deployArgs.PodArgs.MemSoftLimit, resource.DecimalSI)
	}
	if deployArgs.PodArgs.CPULimit != 0 {
		resources.Limits[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(deployArgs.PodArgs.CPULimit*1000), resource.DecimalSI)
	}
	if deployArgs.PodArgs.CPUSoftLimit != 0 {
		resources.Requests[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(deployArgs.PodArgs.CPUSoftLimit*1000), resource.DecimalSI)
	}
	if deployArgs.PodArgs.Nvidia != 0 {
		resources.Limits["nvidia.com/gpu"] = *(resource.NewQuantity(deployArgs.PodArgs.Nvidia, resource.DecimalSI))
	}
	if deployArgs.PodArgs.AMD != 0 {
		resources.Limits["amd.com/gpu"] = *(resource.NewQuantity(deployArgs.PodArgs.AMD, resource.DecimalSI))
	}
	//Copy the existing label
	label := make(map[string]string)
	for k, v := range deployArgs.PodArgs.Label {
		label[k] = v
	}
	//Update the deployment ID label
	label["deploymentId"] = shortuuid.New()
	//Let Kubernetes handle any other invalid entries
	deployment := v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "app/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   deployArgs.DeploymentName,
			Labels: label,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &deployArgs.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: deployArgs.PodArgs.Label},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: label},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      deployArgs.PodArgs.Name,
							Image:     deployArgs.PodArgs.Image,
							Args:      deployArgs.PodArgs.Arguments,
							Command:   deployArgs.PodArgs.Entrypoint,
							Ports:     containerPortLocalToK8s(deployArgs.PodArgs.ExposePorts),
							Env:       envLocalToK8s(deployArgs.PodArgs.Environment),
							Resources: resources,
						},
					},
					ImagePullSecrets: arrayToSecret(secrets),
				},
			},
			Strategy:             v1.DeploymentStrategy{Type: deployArgs.Strategy},
			RevisionHistoryLimit: util.Int32Ptr(0),
		},
		Status: v1.DeploymentStatus{},
	}
	if deployArgs.PodArgs.PullPolicy != "" {
		deployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = deployArgs.PodArgs.PullPolicy
	}
	_, err := kuber.AppsV1().Deployments("default").Update(context.Background(), &deployment, metav1.UpdateOptions{})
	if err != nil {
		log.Error.Println("Update deployment " + deployArgs.DeploymentName + " failed")
		log.Error.Println(err)
	}
	log.Info.Printf("Deployment %s updated\n", deployArgs.DeploymentName)
	return err
}

//Creates a service in Kubernetes
//In order for external IPs to connect to deployments, you'll need services.
//Simply exposing the ports from the pods are not enough. It only allows other pods to communicate with each other
func KCreateService(serviceArgs types.KServiceArgs) (*types.KServiceArgs, error) {
	//You can create a service with no exposed ports, but it's a waste of resources
	if len(serviceArgs.Ports) == 0 {
		return nil, errors.New("no exposed ports")
	}
	kuber := getKuber()
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceArgs.Name,
		},
		Spec: corev1.ServiceSpec{
			Ports:        servicePortLocalToK8s(serviceArgs.Ports),
			Selector:     serviceArgs.Selector,
			Type:         serviceArgs.Type,
			ExternalName: serviceArgs.ExternalName,
		},
		Status: corev1.ServiceStatus{},
	}
	newService, err := kuber.CoreV1().Services("default").Create(context.Background(), &service, metav1.CreateOptions{})
	if err != nil {
		log.Error.Println("Create service " + serviceArgs.Name + " failed")
		log.Error.Println(err)
		return nil, err
	}
	deployedService := types.KServiceArgs{
		Name:            newService.Name,
		Ports:           servicePortK8sToLocal(newService.Spec.Ports),
		Selector:        newService.Spec.Selector,
		ResourceVersion: newService.ResourceVersion,
		Type:            newService.Spec.Type,
		ClusterIP:       newService.Spec.ClusterIP,
		ExternalName:    newService.Spec.ExternalName,
	}
	log.Info.Printf("Service %s created\n", serviceArgs.Name)
	vars.Services.Store(serviceArgs.Name, nil)
	return &deployedService, nil
}

//Deletes a service in Kubernetes
func KDeleteService(serviceName string) error {
	kuber := getKuber()
	err := kuber.CoreV1().Services("default").Delete(context.Background(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		log.Error.Println("Delete service " + serviceName + " failed")
		log.Error.Println(err)
	}
	log.Info.Printf("Service %s deleted\n", serviceName)
	vars.Services.Delete(serviceName)
	return err
}

//Updates a service in Kubernetes
func KUpdateService(serviceArgs types.KServiceArgs) (*types.KServiceArgs, error) {
	kuber := getKuber()
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            serviceArgs.Name,
			ResourceVersion: serviceArgs.ResourceVersion,
		},
		Spec: corev1.ServiceSpec{
			Ports:        servicePortLocalToK8s(serviceArgs.Ports),
			Selector:     serviceArgs.Selector,
			Type:         serviceArgs.Type,
			ExternalName: serviceArgs.ExternalName,
			ClusterIP:    serviceArgs.ClusterIP,
		},
		Status: corev1.ServiceStatus{},
	}
	newService, err := kuber.CoreV1().Services("default").Update(context.Background(), &service, metav1.UpdateOptions{})
	if err != nil {
		log.Error.Println("Update service " + serviceArgs.Name + " failed")
		log.Error.Println(err)
		return nil, err
	}
	//Trim down all of the unnecessary fields
	deployedService := types.KServiceArgs{
		Name:            newService.Name,
		Ports:           servicePortK8sToLocal(newService.Spec.Ports),
		Selector:        newService.Spec.Selector,
		ResourceVersion: newService.ResourceVersion,
		Type:            newService.Spec.Type,
		ClusterIP:       newService.Spec.ClusterIP,
		ExternalName:    newService.Spec.ExternalName,
	}
	log.Info.Printf("Service %s updated\n", serviceArgs.Name)
	return &deployedService, nil
}

//Starts a job in Kubernetes
func KRunJob(jobArgs types.KJobArgs, secrets []string) error {
	kuber := getKuber()
	jobSpec, err := buildJobSpec(jobArgs, secrets, true)
	if err != nil {
		return err
	}
	job := batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: jobArgs.JobName},
		Spec:       *jobSpec,
	}
	//Let Kubernetes handle any other invalid entries
	_, err = kuber.BatchV1().Jobs("default").Create(context.Background(), &job, metav1.CreateOptions{})
	//The Go library for Kubernetes is based on REST
	//The function returns immediately after the container is made but Kubernetes doesn't care if the container starts correctly
	//An error only occurs at the REST level (Input format errors, connection failure, etc.)
	//It does not return an error if the container configurations are incorrect (Invalid image, container crashes, etc.)
	//To check if the container is deployed correctly. You must inspect the job

	//If we need to know all currently deployed job, we can simply ask Kubernetes instead of a database
	if err != nil {
		log.Error.Println("Create job " + jobArgs.JobName + " failed")
		log.Error.Println(err)
	}
	log.Info.Printf("Job %s created\n", jobArgs.JobName)
	vars.Jobs.Store(jobArgs.JobName, nil)
	return err
}

func KDeleteJob(jobName string) error {
	kuber := getKuber()
	err := kuber.BatchV1().Jobs("default").Delete(context.Background(), jobName, metav1.DeleteOptions{})
	if err != nil {
		log.Error.Println("Delete job " + jobName + " failed")
		log.Error.Println(err)
	}
	log.Info.Printf("Job %s deleted\n", jobName)
	vars.Jobs.Delete(jobName)
	return err
}

func KRunCronjob(cronArgs types.KCronjobArgs, secrets []string) error {
	kuber := getKuber()
	jobSpec, err := buildJobSpec(cronArgs.JobArgs, secrets, false)
	if err != nil {
		return err
	}
	cronjob := batchv1b.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: "batch/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: cronArgs.JobName},
		Spec: batchv1b.CronJobSpec{
			Schedule: cronArgs.Cron,
			JobTemplate: batchv1b.JobTemplateSpec{
				Spec: *jobSpec,
			},
			SuccessfulJobsHistoryLimit: util.Int32Ptr(1),
		},
	}
	//Let Kubernetes handle any other invalid entries
	_, err = kuber.BatchV1beta1().CronJobs("default").Create(context.Background(), &cronjob, metav1.CreateOptions{})
	//The Go library for Kubernetes is based on REST
	//The function returns immediately after the container is made but Kubernetes doesn't care if the container starts correctly
	//An error only occurs at the REST level (Input format errors, connection failure, etc.)
	//It does not return an error if the container configurations are incorrect (Invalid image, container crashes, etc.)
	//To check if the container is deployed correctly. You must inspect the job

	//If we need to know all currently deployed job, we can simply ask Kubernetes instead of a database
	if err != nil {
		log.Error.Println("Create cronjob " + cronArgs.JobName + " failed")
		log.Error.Println(err)
	}
	log.Info.Printf("CronJob %s created\n", cronArgs.JobName)
	vars.CronJobs.Store(cronArgs.JobName, nil)
	return err
}

func KDeleteCronJob(cronjobName string) error {
	kuber := getKuber()
	err := kuber.BatchV1beta1().CronJobs("default").Delete(context.Background(), cronjobName, metav1.DeleteOptions{})
	if err != nil {
		log.Error.Println("Delete cronjob " + cronjobName + " failed")
		log.Error.Println(err)
	}
	log.Info.Printf("CronJob %s deleted\n", cronjobName)
	vars.CronJobs.Delete(cronjobName)
	return err
}

//Lists all deployments
func KListDeployment() ([]types.KDeployArgs, error) {
	kuber := getKuber()
	list, err := kuber.AppsV1().Deployments("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	deployments := make([]types.KDeployArgs, 0)
	//Trim down all of the unnecessary fields.
	for _, item := range list.Items {
		deployment, err := KGetDeployment(item.Name)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, *deployment)
	}
	return deployments, nil
}

//Gets a single deployment given its name
func KGetDeployment(deploymentName string) (*types.KDeployArgs, error) {
	kuber := getKuber()
	kdeploy, err := kuber.AppsV1().Deployments("default").Get(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	deployment := types.KDeployArgs{
		PodArgs:        fetchPodArgs(kdeploy.Spec.Template.Spec, kdeploy.Spec.Template.ObjectMeta.Labels),
		DeploymentName: kdeploy.Name,
		Replicas:       *kdeploy.Spec.Replicas,
		Strategy:       kdeploy.Spec.Strategy.Type,
	}
	return &deployment, nil
}

//Lists all deployed services
func KListService() ([]types.KServiceArgs, error) {
	kuber := getKuber()
	list, err := kuber.CoreV1().Services("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	services := make([]types.KServiceArgs, 0)
	for _, k8sService := range list.Items {
		services = append(services, types.KServiceArgs{
			Name:            k8sService.Name,
			Ports:           servicePortK8sToLocal(k8sService.Spec.Ports),
			Selector:        k8sService.Spec.Selector,
			ResourceVersion: k8sService.ResourceVersion,
			Type:            k8sService.Spec.Type,
			ClusterIP:       k8sService.Spec.ClusterIP,
			ExternalName:    k8sService.Spec.ExternalName,
		})
	}
	return services, nil
}

//Gets a single service given its name
func KGetService(serviceName string) (*types.KServiceArgs, error) {
	kuber := getKuber()
	kservice, err := kuber.CoreV1().Services("default").Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	service := types.KServiceArgs{
		Name:            kservice.Name,
		Ports:           servicePortK8sToLocal(kservice.Spec.Ports),
		Selector:        kservice.Spec.Selector,
		ResourceVersion: kservice.ResourceVersion,
		Type:            kservice.Spec.Type,
		ClusterIP:       kservice.Spec.ClusterIP,
		ExternalName:    kservice.Spec.ExternalName,
	}
	return &service, nil
}

func KListJob() ([]types.KJobArgs, error) {
	kuber := getKuber()
	list, err := kuber.BatchV1().Jobs("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	jobs := make([]types.KJobArgs, 0)
	//Trim down all of the unnecessary fields.
	for _, item := range list.Items {
		job, err := KGetJob(item.Name)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *job)
	}
	return jobs, nil
}

func KGetJob(jobName string) (*types.KJobArgs, error) {
	kuber := getKuber()
	kJob, err := kuber.BatchV1().Jobs("default").Get(context.Background(), jobName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	job := fetchJobArgs(kJob.Spec, kJob.Labels)
	job.JobName = kJob.Name
	return &job, nil
}

func KListCronjob() ([]types.KCronjobArgs, error) {
	kuber := getKuber()
	list, err := kuber.BatchV1beta1().CronJobs("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cronjobs := make([]types.KCronjobArgs, 0)
	//Trim down all of the unnecessary fields.
	for _, item := range list.Items {
		cron, err := KGetCronjob(item.Name)
		if err != nil {
			return nil, err
		}
		cronjobs = append(cronjobs, *cron)
	}
	return cronjobs, nil
}

func KGetCronjob(cronjobName string) (*types.KCronjobArgs, error) {
	kuber := getKuber()
	kCronjob, err := kuber.BatchV1beta1().CronJobs("default").Get(context.Background(), cronjobName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	jobArgs := fetchJobArgs(kCronjob.Spec.JobTemplate.Spec, kCronjob.Spec.JobTemplate.Spec.Template.ObjectMeta.Labels)
	cronjob := types.KCronjobArgs{
		JobArgs: jobArgs,
		JobName: kCronjob.Name,
		Cron:    kCronjob.Spec.Schedule,
	}
	return &cronjob, nil
}

func KCreateConfigMap(configMap types.KConfigMap) error {
	kuber := getKuber()
	kconfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configMap.Name,
		},
		Data:       configMap.Data,
		BinaryData: configMap.BinaryData,
	}
	_, err := kuber.CoreV1().ConfigMaps("default").Create(context.Background(), kconfigMap, metav1.CreateOptions{})
	return err
}

func KDeleteConfigMap(configMapName string) error {
	kuber := getKuber()
	err := kuber.CoreV1().ConfigMaps("default").Delete(context.Background(), configMapName, metav1.DeleteOptions{})
	return err
}

func KListConfigMap() ([]types.KConfigMap, error) {
	kuber := getKuber()
	list, err := kuber.CoreV1().ConfigMaps("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	configMaps := make([]types.KConfigMap, 0)
	for _, item := range list.Items {
		config := types.KConfigMap{
			Name:       item.Name,
			Data:       item.Data,
			BinaryData: item.BinaryData,
		}
		configMaps = append(configMaps, config)
	}
	return configMaps, nil
}

//Returns a deployment watcher
//The channel first returns all current deployments, then reports and changes to deployments
func KGetDeploymentWatcher() (watch.Interface, error) {
	kuber := getKuber()
	return kuber.AppsV1().Deployments("default").Watch(context.Background(), metav1.ListOptions{
		TimeoutSeconds: util.Int64Ptr(86400),
	})
}

//Returns a pod watcher. A label selector is recommended to limit the watch scope
func KGetPodWatcher(labelSelector map[string]string) (watch.Interface, error) {
	kuber := getKuber()
	return kuber.CoreV1().Pods("default").Watch(context.Background(), metav1.ListOptions{
		LabelSelector:  labels.Set(labelSelector).String(),
		TimeoutSeconds: util.Int64Ptr(86400),
	})
}

//Returns a cronjob watcher
//Watch calls first returns all current objects. Then it returns any changed/added/removed object
func KGetCronjobWatcher() (watch.Interface, error) {
	kuber := getKuber()
	return kuber.BatchV1beta1().CronJobs("default").Watch(context.Background(), metav1.ListOptions{
		TimeoutSeconds: util.Int64Ptr(86400),
	})
}

//Returns a job watcher. This shows all information about the job, including whatever it is part of a cronjob
//Watch calls first returns all current objects. Then it returns any changed/added/removed object
func KGetJobWatcher() (watch.Interface, error) {
	kuber := getKuber()
	return kuber.BatchV1().Jobs("default").Watch(context.Background(), metav1.ListOptions{
		TimeoutSeconds: util.Int64Ptr(86400),
	})
}

//Convert Kubernetes API PodSpec back to local type KPodArgs
func fetchPodArgs(kPodSpec corev1.PodSpec, labels map[string]string) types.KPodArgs {
	podArgs := types.KPodArgs{
		Image:         kPodSpec.Containers[0].Image,
		Label:         labels,
		Name:          kPodSpec.Containers[0].Name,
		ExposePorts:   containerPortK8sToLocal(kPodSpec.Containers[0].Ports),
		Arguments:     kPodSpec.Containers[0].Args,
		Entrypoint:    kPodSpec.Containers[0].Command,
		Environment:   envK8sToLocal(kPodSpec.Containers[0].Env),
		MemLimit:      kPodSpec.Containers[0].Resources.Limits.Memory().Value(),
		MemSoftLimit:  kPodSpec.Containers[0].Resources.Requests.Memory().Value(),
		PullPolicy:    kPodSpec.Containers[0].ImagePullPolicy,
		CPULimit:      float64(kPodSpec.Containers[0].Resources.Limits.Cpu().MilliValue()) / 1000,
		CPUSoftLimit:  float64(kPodSpec.Containers[0].Resources.Requests.Cpu().MilliValue()) / 1000,
		RestartPolicy: kPodSpec.RestartPolicy,
	}
	volumes := make([]types.KVolume, 0)
	for _, hostMountPoint := range kPodSpec.Volumes {
		mountName := hostMountPoint.Name
		if hostMountPoint.HostPath != nil {
			for _, containerMountPoint := range kPodSpec.Containers[0].VolumeMounts {
				if containerMountPoint.Name == mountName {
					volume := types.KVolume{
						MountPoint: containerMountPoint.MountPath,
						ReadOnly:   containerMountPoint.ReadOnly,
						Type:       types.TypeHostPath,
						HostPath: &types.HostPath{
							Path:         hostMountPoint.HostPath.Path,
							MountType:    *hostMountPoint.HostPath.Type,
							NodeHostname: kPodSpec.NodeSelector["kubernetes.io/hostname"],
						},
					}
					volumes = append(volumes, volume)
				}
			}
		} else if hostMountPoint.AWSElasticBlockStore != nil {
			for _, containerMountPoint := range kPodSpec.Containers[0].VolumeMounts {
				if containerMountPoint.Name == mountName {
					volume := types.KVolume{
						MountPoint: containerMountPoint.MountPath,
						ReadOnly:   containerMountPoint.ReadOnly,
						Type:       types.TypeAWSEBS,
						AwsEbs: &types.AwsEbs{
							VolumeID:  hostMountPoint.AWSElasticBlockStore.VolumeID,
							FSType:    hostMountPoint.AWSElasticBlockStore.FSType,
							Partition: hostMountPoint.AWSElasticBlockStore.Partition,
						},
					}
					volumes = append(volumes, volume)
				}
			}
		} else if hostMountPoint.NFS != nil {
			for _, containerMountPoint := range kPodSpec.Containers[0].VolumeMounts {
				if containerMountPoint.Name == mountName {
					volume := types.KVolume{
						MountPoint: containerMountPoint.MountPath,
						ReadOnly:   containerMountPoint.ReadOnly,
						Type:       types.TypeNFS,
						NFS: &types.NFS{
							Server: hostMountPoint.NFS.Server,
							Path:   hostMountPoint.NFS.Path,
						},
					}
					volumes = append(volumes, volume)
				}
			}
		} else if hostMountPoint.EmptyDir != nil {
			for _, containerMountPoint := range kPodSpec.Containers[0].VolumeMounts {
				if containerMountPoint.Name == mountName {
					volume := types.KVolume{
						MountPoint: containerMountPoint.MountPath,
						ReadOnly:   containerMountPoint.ReadOnly,
						Type:       types.TypeEmptyDir,
						EmptyDir: &types.EmptyDir{
							StorageMedium: hostMountPoint.EmptyDir.Medium,
						},
					}
					volumes = append(volumes, volume)
				}
			}
		} else if hostMountPoint.ConfigMap != nil {
			for _, containerMountPoint := range kPodSpec.Containers[0].VolumeMounts {
				if containerMountPoint.Name == mountName {
					volume := types.KVolume{
						MountPoint: containerMountPoint.MountPath,
						ReadOnly:   containerMountPoint.ReadOnly,
						Type:       types.TypeConfigMap,
						ConfigMap: &types.ConfigMapVolume{
							Name: hostMountPoint.ConfigMap.Name,
							Mode: hostMountPoint.ConfigMap.DefaultMode,
						},
					}
					volumes = append(volumes, volume)
				}
			}
		}
	}
	podArgs.Volumes = volumes
	return podArgs
}

func fetchJobArgs(kJobSpec batchv1.JobSpec, labels map[string]string) types.KJobArgs {
	jobArgs := types.KJobArgs{
		PodArgs: fetchPodArgs(kJobSpec.Template.Spec, labels),
	}
	if kJobSpec.BackoffLimit != nil {
		jobArgs.BackOffLimit = *kJobSpec.BackoffLimit
	}
	return jobArgs
}

//Build Kubernetes API PodSpec from local type KPodArgs
func buildPodSpec(name string, podArgs types.KPodArgs, secrets []string) (*corev1.PodSpec, error) {
	var volumes = make([]corev1.Volume, 0)
	var mountPoints = make([]corev1.VolumeMount, 0)
	var nodeSelector string
	for _, vol := range podArgs.Volumes {
		//Generate a name for the volume
		//When creating a deployment, we first define a volume with a name
		//Then the pods uses that name to mount the volume.
		//Names must match this regex /[a-z0-9]([-a-z0-9]*[a-z0-9])?/
		volName := strings.ToLower(fmt.Sprintf("%s-%s", name, shortuuid.New()))
		// Limitations to hostpath volumes:
		// Pods must be deployed to only one specific node
		// The pods deployed will only mount to that node's host directory
		// If you have more than one hostpath volume, they must use the same node
		if vol.Type == types.TypeHostPath {
			if vol.HostPath == nil {
				return nil, errors.New("did not specify volume info")
			}
			if strings.TrimSpace(vol.HostPath.NodeHostname) == "" {
				return nil, errors.New("node hostname must not be empty")
			} else if nodeSelector == "" {
				nodeSelector = vol.HostPath.NodeHostname
			} else if nodeSelector != vol.HostPath.NodeHostname {
				return nil, errors.New("hostname must be the same across all hostpath volumes")
			}
			volumes = append(volumes, corev1.Volume{
				Name: volName,
				VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{
					Path: vol.HostPath.Path,
					Type: &vol.HostPath.MountType,
				}},
			})
			mountPoints = append(mountPoints, corev1.VolumeMount{
				Name:      volName,
				ReadOnly:  vol.ReadOnly,
				MountPath: vol.MountPoint,
			})
		} else if vol.Type == types.TypeAWSEBS {
			if vol.AwsEbs == nil {
				return nil, errors.New("did not specify volume info")
			}
			volumes = append(volumes, corev1.Volume{
				Name: volName,
				VolumeSource: corev1.VolumeSource{AWSElasticBlockStore: &corev1.AWSElasticBlockStoreVolumeSource{
					VolumeID:  vol.AwsEbs.VolumeID,
					FSType:    vol.AwsEbs.FSType,
					Partition: vol.AwsEbs.Partition,
					ReadOnly:  vol.ReadOnly,
				}},
			})
			mountPoints = append(mountPoints, corev1.VolumeMount{
				Name:      volName,
				ReadOnly:  vol.ReadOnly,
				MountPath: vol.MountPoint,
			})
		} else if vol.Type == types.TypeNFS {
			if vol.NFS == nil {
				return nil, errors.New("did not specify volume info")
			}
			volumes = append(volumes, corev1.Volume{
				Name: volName,
				VolumeSource: corev1.VolumeSource{NFS: &corev1.NFSVolumeSource{
					Server:   vol.NFS.Server,
					Path:     vol.NFS.Path,
					ReadOnly: vol.ReadOnly,
				}},
			})
			mountPoints = append(mountPoints, corev1.VolumeMount{
				Name:      volName,
				ReadOnly:  vol.ReadOnly,
				MountPath: vol.MountPoint,
			})
		} else if vol.Type == types.TypeEmptyDir {
			if vol.EmptyDir == nil {
				return nil, errors.New("did not specify volume info")
			}
			emptyDir := corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: vol.EmptyDir.StorageMedium,
			}}
			if vol.EmptyDir.SizeLimit != nil {
				emptyDir.EmptyDir.SizeLimit = resource.NewQuantity(*vol.EmptyDir.SizeLimit, resource.DecimalSI)
			}
			volumes = append(volumes, corev1.Volume{
				Name:         volName,
				VolumeSource: emptyDir,
			})
			mountPoints = append(mountPoints, corev1.VolumeMount{
				Name:      volName,
				ReadOnly:  vol.ReadOnly,
				MountPath: vol.MountPoint,
			})
		} else if vol.Type == types.TypeConfigMap {
			if vol.ConfigMap == nil {
				return nil, errors.New("did not specify volume info")
			}
			volumes = append(volumes, corev1.Volume{
				Name: volName,
				VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: vol.ConfigMap.Name,
					},
					DefaultMode: vol.ConfigMap.Mode,
				}},
			})
			mountPoints = append(mountPoints, corev1.VolumeMount{
				Name:      volName,
				ReadOnly:  vol.ReadOnly,
				MountPath: vol.MountPoint,
			})
		} else {
			//Raise error if did not specify the volume type
			return nil, errors.New("unknown/did not specify volume type")
		}
	}

	resources := corev1.ResourceRequirements{
		Limits:   make(corev1.ResourceList),
		Requests: make(corev1.ResourceList),
	}
	//Only fill in information if it exists. Otherwise Kubernetes will put 0 for the limits instead of automatic.
	if podArgs.MemLimit != 0 {
		resources.Limits[corev1.ResourceMemory] = *resource.NewQuantity(podArgs.MemLimit, resource.DecimalSI)
	}
	if podArgs.MemSoftLimit != 0 {
		resources.Requests[corev1.ResourceMemory] = *resource.NewQuantity(podArgs.MemSoftLimit, resource.DecimalSI)
	}
	if podArgs.CPULimit != 0 {
		resources.Limits[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(podArgs.CPULimit*1000), resource.DecimalSI)
	}
	if podArgs.CPUSoftLimit != 0 {
		resources.Requests[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(podArgs.CPUSoftLimit*1000), resource.DecimalSI)
	}
	if podArgs.Nvidia != 0 {
		resources.Limits["nvidia.com/gpu"] = *(resource.NewQuantity(podArgs.Nvidia, resource.DecimalSI))
	}
	if podArgs.AMD != 0 {
		resources.Limits["amd.com/gpu"] = *(resource.NewQuantity(podArgs.AMD, resource.DecimalSI))
	}
	podSpec := corev1.PodSpec{
		Volumes: volumes,
		Containers: []corev1.Container{
			{
				Name:         podArgs.Name,
				Image:        podArgs.Image,
				Args:         podArgs.Arguments,
				Command:      podArgs.Entrypoint,
				Ports:        containerPortLocalToK8s(podArgs.ExposePorts),
				Env:          envLocalToK8s(podArgs.Environment),
				Resources:    resources,
				VolumeMounts: mountPoints,
			},
		},
		ImagePullSecrets: arrayToSecret(secrets),
		RestartPolicy:    podArgs.RestartPolicy,
	}
	//Deployments that use hostpath volumes must select a node to deploy on as they are not multi-node compatible.
	if nodeSelector != "" {
		podSpec.NodeSelector = map[string]string{
			"kubernetes.io/hostname": nodeSelector,
		}
	}
	if podArgs.PullPolicy != "" {
		podSpec.Containers[0].ImagePullPolicy = podArgs.PullPolicy
	}
	if podArgs.Accelerator != "" {
		podSpec.NodeSelector = map[string]string{"accelerator": podArgs.Accelerator}
	}
	return &podSpec, nil
}

//Build Kubernetes API JobSpec from local type KJobArgs
func buildJobSpec(jobArgs types.KJobArgs, secrets []string, manualSelector bool) (*batchv1.JobSpec, error) {
	//CPUTime to live is constant.
	podSpec, err := buildPodSpec(jobArgs.JobName, jobArgs.PodArgs, secrets)
	if err != nil {
		return nil, err
	}
	jobspec := batchv1.JobSpec{
		BackoffLimit: &jobArgs.BackOffLimit,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: jobArgs.PodArgs.Label},
			Spec:       *podSpec,
		},
		TTLSecondsAfterFinished: util.Int32Ptr(30),
		ManualSelector:          &manualSelector,
	}
	if manualSelector {
		jobspec.Selector = &metav1.LabelSelector{MatchLabels: jobArgs.PodArgs.Label}
	}
	return &jobspec, nil
}

//Convert types part 1
//Kubernetes interfaces while powerful, many of them are unused in this project.
//These functions are here to make things easier by trimming out everything unnecessary.
//Local type to Kubernetes container port settings
func containerPortLocalToK8s(ports []types.ExposePort) []corev1.ContainerPort {
	containerPort := make([]corev1.ContainerPort, 0)
	for _, port := range ports {
		cp := corev1.ContainerPort{
			HostPort:      int32(port.HostPort),
			ContainerPort: int32(port.ContainerPort),
		}
		if port.Protocol == "" {
			cp.Protocol = corev1.ProtocolTCP
		} else {
			//Upper case it, otherwise Kubernetes doesn't like it
			cp.Protocol = corev1.Protocol(strings.ToUpper(string(port.Protocol)))
		}
		containerPort = append(containerPort, cp)
	}
	return containerPort
}

//Convert types part 2
//Kubernetes container port settings to local type
func containerPortK8sToLocal(kPorts []corev1.ContainerPort) []types.ExposePort {
	exposedPort := make([]types.ExposePort, 0)
	for _, kPort := range kPorts {
		exposedPort = append(exposedPort, types.ExposePort{
			HostPort:      uint16(kPort.HostPort),
			ContainerPort: uint16(kPort.ContainerPort),
			//Lower case it to maintain consistency
			Protocol: types.Protocol(strings.ToLower(string(kPort.Protocol))),
		})
	}
	return exposedPort
}

//Convert types part 3
//Local type to Kubernetes environment
func envLocalToK8s(envMap []types.Environment) []corev1.EnvVar {
	kEnvMap := make([]corev1.EnvVar, 0)
	for _, env := range envMap {
		kEnvMap = append(kEnvMap, corev1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return kEnvMap
}

//Convert types part 4
//Kubernetes environment to local type
func envK8sToLocal(envMap []corev1.EnvVar) []types.Environment {
	mEvnMap := make([]types.Environment, 0)
	for _, env := range envMap {
		mEvnMap = append(mEvnMap, types.Environment{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return mEvnMap
}

//Convert types part 5
//LocalObjectReference is just a struct with one string field.
//Array to LocalObjectReference
func arrayToSecret(secret []string) []corev1.LocalObjectReference {
	kSecret := make([]corev1.LocalObjectReference, 0)
	if secret == nil {
		return kSecret
	}
	for _, s := range secret {
		kSecret = append(kSecret, corev1.LocalObjectReference{Name: s})
	}
	return kSecret
}

//Convert types part 6
//ServicePort apparently is not equal to ContainerPort
//Local type to Kubernetes service port
func servicePortLocalToK8s(ports []types.KServicePort) []corev1.ServicePort {
	servicePorts := make([]corev1.ServicePort, 0)
	for _, port := range ports {
		sp := corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Host,
			TargetPort: intstr.FromInt(int(port.TargetPort)),
			NodePort:   port.NodePort,
		}
		if port.Protocol == "" {
			sp.Protocol = corev1.ProtocolTCP
		} else {
			//Upper case it, otherwise Kubernetes doesn't like it
			sp.Protocol = corev1.Protocol(strings.ToUpper(string(port.Protocol)))
		}
		servicePorts = append(servicePorts, sp)
	}
	return servicePorts
}

//Convert types part 7
//Kubernetes service port setting to local type
func servicePortK8sToLocal(servicePort []corev1.ServicePort) []types.KServicePort {
	mp := make([]types.KServicePort, 0)
	for _, sp := range servicePort {
		mp = append(mp, types.KServicePort{
			Name:       sp.Name,
			Host:       sp.Port,
			TargetPort: sp.TargetPort.IntVal,
			NodePort:   sp.NodePort,
			//Lower case it to maintain consistency
			Protocol: types.Protocol(strings.ToLower(string(sp.Protocol))),
		})
	}
	return mp
}

//Mutex to collect Kubernetes config only once
var configMutex sync.Once
var kClient *kubernetes.Clientset

//Get Kubernetes configurations
//This includes credentials, the address the client should connect to, etc.
func getKuber() *kubernetes.Clientset {
	configMutex.Do(func() {
		if vars.GetKuberConfigPath() == "" {
			log.Fatal.Panicln("Kubernetes config filepath not defined!")
		}
		config, err := clientcmd.BuildConfigFromFlags("", vars.GetKuberConfigPath())
		if err != nil {
			log.Fatal.Println("Cannot read Kubernetes config")
			log.Fatal.Panicln(err)
		}
		kClient, err = kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatal.Println("Error occurred generating client for Kubernetes")
			log.Fatal.Panicln(err)
		}
	})
	return kClient
}
