package monitor

import (
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"osmoticframework/controller/alert"
	"osmoticframework/controller/api/impl/request"
	"osmoticframework/controller/log"
	"osmoticframework/controller/types"
	"sync"
	"time"
)

//Monitors Kubernetes container status via Kubernetes API watch call
//Reports any anomaly to the alert channel

//Kubernetes provides events which reports events that happened on all Kubernetes resources
//You can use `kubectl get events` to get all events (use --field-selector involvedObject.kind= to filter specific resource types)
//However they are extremely unreliable as Kubernetes automatically merge similar events together
//We lost a lot of real time info and missed a lot of events

//An enhancement proposal is made last year but it is still in alpha and not merged to release (03/2021 as of yet)
//https://github.com/kubernetes/enhancements/issues/1440

//The only solution is to directly watch changes to the resource object
// which sadly the resource object does not report all events
//...and it makes everything needlessly complicated

//Monitors deployment health, specifically the pods deployed by the ReplicaSet from the deployment
//Reports any crashes to the alert channel
func DeploymentWatch() {
	//Stores the pod watchers mapped to the deployment name
	var podWatchers = make(map[string]*PodWatcher)
	//Stores the current generation number of a deployment, mapped to the deployment name
	var generations = make(map[string]int64)
	for {
		watcher, err := request.KGetDeploymentWatcher()
		if err != nil {
			log.Warn.Println("Failed to obtain event watcher from Kubernetes. Retrying in 5 seconds")
			log.Warn.Println(err)
			time.Sleep(time.Second * 5)
			continue
		}
		log.Info.Println("Started Kubernetes deployment watcher")
		for _event := range watcher.ResultChan() {
			eventType := _event.Type
			event := _event.Object.(*v1.Deployment)
			deploymentName := event.Name
			//Deployments themselves do not create Pods, nor do they monitor pods
			//Therefore, watching events of deployment will not help. Instead, we monitor the changes to deployments directly
			//This watcher returns the deployment added/removed/changed in question
			switch eventType {
			case watch.Added:
				//Do not create a new pod watcher if one already exists
				if _, ok := podWatchers[deploymentName]; ok {
					continue
				}
				//Usually refer to new deployments
				podWatch := PodWatcher{}
				//Start the monitoring thread and store the pointer to memory
				podWatch.Monitor(deploymentName, event.ObjectMeta.Labels["deploymentId"])
				podWatchers[deploymentName] = &podWatch
				generations[deploymentName] = event.ObjectMeta.Generation
			case watch.Deleted:
				//Removal of deployments
				podWatch, ok := podWatchers[deploymentName]
				if !ok {
					log.Warn.Printf("Cannot find watchers from deleted deployment %s! This may lead to wasted resources\n", deploymentName)
				} else {
					//Stop the watcher and remove everything from memory
					podWatch.Stop()
				}
				delete(podWatchers, deploymentName)
				delete(generations, deploymentName)
			case watch.Modified:
				//Modified events include configuration changes and events that happened to a deployment
				//We use the generation number to keep track of updated deployments
				if event.Generation == generations[deploymentName] {
					continue
				}
				//If the generation number has changed, this means the deployment configuration has changed and it's not an internal event.
				//Update the generation number
				generations[deploymentName] = event.ObjectMeta.Generation
				//Stop the current watcher
				podWatchers[deploymentName].Stop()
				//Create a new one
				podWatch := PodWatcher{}
				//Start the monitoring thread and store the pointer to memory
				podWatch.Monitor(deploymentName, event.Labels["deploymentId"])
				podWatchers[deploymentName] = &podWatch
			default:
				//Ignore all other types of events
			}
		}
		log.Warn.Println("Kubernetes has hung up. Restarting deployment watcher")
	}
}

//CronJobs events shows when the cronjob executes
//But the cronjob object does not reveal anything about its status
//This thread is only used to monitor what cronjobs are created and deleted

//Stores all current cronjob names as an array
//Use maps as they are faster and searching in array is slow
var currentCronJobs sync.Map

func CronjobWatch() {
	for {
		watcher, err := request.KGetCronjobWatcher()
		if err != nil {
			log.Warn.Println("Failed to obtain event watcher from Kubernetes. Retrying in 5 seconds")
			log.Warn.Println(err)
			time.Sleep(time.Second * 5)
			continue
		}
		log.Info.Println("Started Kubernetes cronjob event watcher")
		for _event := range watcher.ResultChan() {
			eventType := _event.Type
			event := _event.Object.(*v1beta1.CronJob)
			switch eventType {
			//We only care when a cronjob is created or deleted. All cronjob events will be monitored in the job watcher instead
			case watch.Added:
				currentCronJobs.Store(event.Name, true)
				log.Info.Printf("Cronjob %s created\n", event.Name)
			case watch.Deleted:
				currentCronJobs.Delete(event.Name)
				log.Info.Printf("Cronjob %s deleted\n", event.Name)
			default:
				//Ignore all other types of events
			}
		}
		log.Warn.Println("Kubernetes has hung up. Restarting event watcher")
	}
}

//The Job watch is responsible to monitor all jobs
//CronJobs create Jobs periodically according to its cron schedule
//Fortunately the job object does return its latest status, also whatever it is part of a cronjob.
//Bad news is that we cannot monitor CronJob specific events. Such as when the job associated with the cronjob went missing
func JobWatch() {
	for {
		watcher, err := request.KGetJobWatcher()
		if err != nil {
			log.Warn.Println("Failed to obtain event watcher from Kubernetes. Retrying in 5 seconds")
			log.Warn.Println(err)
			time.Sleep(time.Second * 5)
			continue
		}
		log.Info.Println("Started Kubernetes job event watcher")
		for _event := range watcher.ResultChan() {
			eventType := _event.Type
			event := _event.Object.(*batchv1.Job)
			switch eventType {
			//When a job is added (or the start of the watch call where the listing is treated as added)
			case watch.Added:
				cronjob, ok := isCronjob(*event)
				if ok {
					log.Info.Printf("Job %s of cronjob %s created\n", event.Name, cronjob)
				} else {
					log.Info.Printf("Job %s created\n", event.Name)
				}
			case watch.Deleted:
				cronjob, ok := isCronjob(*event)
				if ok {
					log.Info.Printf("Job %s of cronjob %s deleted\n", event.Name, cronjob)
				} else {
					log.Info.Printf("Job %s deleted\n", event.Name)
				}
			case watch.Modified:
				cronjob, ok := isCronjob(*event)
				if event.Status.Succeeded == 1 {
					if ok {
						log.Info.Printf("Job %s of cronjob %s has finished execution\n", event.Name, cronjob)
						alert.ContainerCrash <- types.ContainerCrashReport{
							AgentId:  "cloud-cronjob",
							ID:       cronjob,
							Name:     event.Name,
							Status:   "exited",
							ExitCode: 0,
						}
					} else {
						log.Info.Printf("Job %s has finished execution\n", event.Name)
						alert.ContainerCrash <- types.ContainerCrashReport{
							AgentId:  "cloud-job",
							ID:       event.Name,
							Status:   "exited",
							ExitCode: 0,
						}
					}
				} else if event.Status.Failed == 1 {
					if ok {
						log.Warn.Printf("Job %s of cronjob %s has failed. Maximum back off limit reached\n", event.Name, cronjob)
						log.Warn.Printf("Cronjob %s has failed scheduled execution\n", cronjob)
						alert.ContainerCrash <- types.ContainerCrashReport{
							AgentId:  "cloud-cronjob",
							ID:       cronjob,
							Name:     event.Name,
							Status:   "error",
							ExitCode: -1,
						}
					} else {
						log.Warn.Printf("Job %s has failed execution. Maximum back off limit reached\n", event.Name)
						alert.ContainerCrash <- types.ContainerCrashReport{
							AgentId:  "cloud-job",
							ID:       event.Name,
							Status:   "error",
							ExitCode: -1,
						}
					}
				} else if event.Status.Active == 1 {
					if ok {
						log.Info.Printf("Job %s of cronjob %s has started\n", event.Name, cronjob)
					} else {
						log.Info.Printf("Job %s has started\n", event.Name)
					}
				}
			default:
				//Ignore all other types of events
			}
		}
		log.Warn.Println("Kubernetes has hung up. Restarting event watcher")
	}
}

func isCronjob(job batchv1.Job) (string, bool) {
	if len(job.OwnerReferences) == 0 {
		return "", false
	}
	for _, ref := range job.OwnerReferences {
		if _, ok := currentCronJobs.Load(ref.Name); ok && ref.Kind == "CronJob" {
			return ref.Name, true
		}
	}
	return "", false
}

//PodWatcher struct
//When initializing one, do not put any values in the struct
type PodWatcher struct {
	watcher watch.Interface
	quit    bool
	u       sync.Mutex
}

//Stop the watcher from other threads
func (pw *PodWatcher) Stop() {
	pw.u.Lock()
	if pw.watcher != nil {
		pw.watcher.Stop()
		pw.quit = true
	}
	pw.u.Unlock()
}

//Pod watcher body
//This starts the watcher go routine and reads the changes to pods and alerts the controller
func (pw *PodWatcher) Monitor(deploymentName, deploymentId string) {
	go func() {
		for {
			//If the quit flag is raised, just exit the go routine
			pw.u.Lock()
			if pw.quit {
				return
			}
			pw.u.Unlock()
			selector := map[string]string{
				"deploymentId": deploymentId,
			}
			var err error
			pw.watcher, err = request.KGetPodWatcher(selector)
			if err != nil {
				log.Warn.Println("Failed to obtain pod watcher from Kubernetes. Retrying in 5 seconds")
				log.Warn.Println(err)
				time.Sleep(time.Second * 5)
				continue
			}
			log.Info.Printf("Started Kubernetes pod watcher for deployment %s\n", deploymentName)
			for _event := range pw.watcher.ResultChan() {
				eventType := _event.Type
				event := _event.Object.(*corev1.Pod)
				var status string
				if len(event.Status.ContainerStatuses) != 0 {
					state := event.Status.ContainerStatuses[0].State
					if state.Running != nil {
						status = "Running"
					} else if state.Terminated != nil {
						status = state.Terminated.Reason
					} else if state.Waiting != nil {
						status = state.Waiting.Reason
					}
				}
				switch eventType {
				case watch.Added:
					//Usually refers to new pod scheduled, or the pods from the start of the watch
					//The pod is scheduled to the queue
					if event.Status.Phase == corev1.PodPending {
						log.Info.Printf("Pod %s of deployment %s is now pending for deployment\n", event.Name, deploymentName)
					}
				case watch.Deleted:
					//Removal of pods
					//Delete events are not guaranteed to appear, as the deployment will stop this thread before this happens
					log.Info.Printf("Pod %s of deployment %s is terminating\n", event.Name, deploymentName)
				case watch.Modified:
					//Change of state in the pod. Such as running, backoff, error, etc.
					if event.Status.Phase == corev1.PodRunning {
						if status == "Running" {
							log.Info.Printf("Pod %s of deployment %s is now running\n", event.Name, deploymentName)
						} else if status == "CrashLoopBackOff" {
							log.Info.Printf("Pod %s of deployment %s has entered backoff state\n", event.Name, deploymentName)
						} else if status == "Error" {
							alert.ContainerCrash <- types.ContainerCrashReport{
								AgentId:  "cloud-deployment",
								ID:       deploymentName,
								Name:     event.Name,
								Image:    event.Spec.Containers[0].Image,
								Status:   "error",
								ExitCode: int(event.Status.ContainerStatuses[0].State.Terminated.ExitCode),
							}
							log.Warn.Printf("Pod %s of deployment %s has crashed\n", event.Name, deploymentName)
						} else if status == "Completed" {
							alert.ContainerCrash <- types.ContainerCrashReport{
								AgentId:  "cloud-deployment",
								ID:       deploymentName,
								Name:     event.Name,
								Image:    event.Spec.Containers[0].Image,
								Status:   "exited",
								ExitCode: 0,
							}
							log.Warn.Printf("Pod %s of deployment %s has exited with code 0. You should use Jobs for one time executed applications", event.Name, deploymentName)
						} else {
							log.Info.Printf("Pod %s of deployment %s status unknown. Attempt dumping all info\n", event.Name, deploymentName)
							log.Info.Printf("%#v", event.Status)
						}
					} else if event.Status.Phase == corev1.PodPending {
						if status == "ContainerCreating" {
							log.Info.Printf("Pod %s of deployment %s is being deployed\n", event.Name, deploymentName)
						} else if status == "ErrImagePull" {
							log.Warn.Printf("Pod %s of deployment %s has failed to pull the required image\n", event.Name, deploymentName)
						} else if status == "ImagePullBackOff" {
							log.Info.Printf("Pod %s of deployment %s has entered back off state from image pulling\n", event.Name, deploymentName)
						}
					} else if event.Status.Phase == corev1.PodSucceeded {
						alert.ContainerCrash <- types.ContainerCrashReport{
							AgentId:  "cloud-deployment",
							ID:       deploymentName,
							Name:     event.Name,
							Image:    event.Spec.Containers[0].Image,
							Status:   "exited",
							ExitCode: 0,
						}
						log.Warn.Printf("Pod %s of deployment %s has exited with code 0. You should use Jobs for one time executed applications\n", event.Name, deploymentName)
					} else if event.Status.Phase == corev1.PodFailed {
						if status == "Evicted" {
							log.Error.Printf("Pod %s of deployment %s has been evicted. Cluster is dangerously low on resources!\n", event.Name, deploymentName)
							alert.ContainerCrash <- types.ContainerCrashReport{
								AgentId:  "cloud-deployment",
								ID:       deploymentName,
								Name:     event.Name,
								Image:    event.Spec.Containers[0].Image,
								Status:   "evicted",
								ExitCode: -1,
							}
						} else {
							log.Info.Printf("Pod %s of deployment %s status unknown. Attempt dumping all info\n", event.Name, deploymentName)
							log.Info.Printf("%#v", event.Status)
						}
					} else if event.Status.Phase == corev1.PodUnknown {
						log.Info.Printf("Pod %s of deployment %s status unknown. Attempt dumping all info\n", event.Name, deploymentName)
						log.Info.Printf("%#v", event.Status)
					}
				default:
					//Ignore the bookmark event as it's is not required in this workflow
				}
			}
			//If quit flag is raised, exit the monitoring routine
			pw.u.Lock()
			if pw.quit {
				return
			}
			pw.u.Unlock()
			//If not this means Kubernetes has hung up and the watcher needs to be reestablished
			log.Warn.Println("Kubernetes has hung up. Restarting pod watcher")
		}
	}()
}
