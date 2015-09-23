package tree_docker

import (
	"github.com/fsouza/go-dockerclient"
	"tree_event"
	"tree_lib"
)

type ContainerInfo struct {
	ID 					string 					`json:"id" toml:"id" yaml:"id"`  // container ID
	Image 				string 					`json:"image" toml:"image" yaml:"image"`   // Container Image Name
	InspectContainer 	*docker.Container 		`json:"inspect" toml:"inspect" yaml:"inspect"`
}

type ImageInfo struct {
	ID 			string              `json:"id" toml:"id" yaml:"id"`
	Name 		string              `json:"name" toml:"name" yaml:"name"`  // Name is a combination of repository:tag from Docker
	Inspect 	*docker.Image       `json:"inspect" toml:"inspect" yaml:"inspect"`
}

var (
	DockerClient 		*docker.Client
	DockerEndpoint = 	"unix:///var/run/docker.sock"
)

func InitDockerClient() (err tree_lib.TreeError) {
	err.From = tree_lib.FROM_INIT_DOCKER_CLIENT
	if DockerClient != nil {
		return
	}
	DockerClient, err.Err = docker.NewClient(DockerEndpoint)
	if !err.IsNull() {
		return
	}
	return
}

func triggerInitEvent() tree_lib.TreeError {
	var (
		err 				tree_lib.TreeError
		dock_containers		[]docker.APIContainers
	)
	err.From = tree_lib.FROM_TRIGGER_INIT_EVENT
	dock_containers, err.Err = DockerClient.ListContainers(docker.ListContainersOptions{All: false})
	if !err.IsNull() {
		return err
	}

	// Triggering event with currently running Docker containers inside
	tree_event.Trigger(&tree_event.Event{Name:tree_event.ON_DOCKER_INIT, LocalVar: dock_containers})
	return err
}


func StartEventListener() (err tree_lib.TreeError) {
	err.From = tree_lib.FROM_START_EVENT_LISTENER
	err = InitDockerClient()
	if !err.IsNull() {
		return
	}

	err = triggerInitEvent()
	if !err.IsNull() {
		return
	}
	// When function will be returned local event for ending docker client will be triggered
	defer tree_event.Trigger(&tree_event.Event{Name: tree_event.ON_DOCKER_END, LocalVar: nil})

	ev := make(chan *docker.APIEvents)
	err.Err = DockerClient.AddEventListener(ev)
	if !err.IsNull() {
		return
	}

	for  {
		err = callEvent(<- ev)
		if !err.IsNull() {
			break
		}
	}

	return
}

func callEvent(event *docker.APIEvents) tree_lib.TreeError {
	var err tree_lib.TreeError
	err.From = tree_lib.FROM_CALL_EVENT
	switch event.Status {
	case "start", "unpouse":
		{
			var (
				dock_inspect	*docker.Container
			)
			dock_inspect, err.Err = DockerClient.InspectContainer(event.ID)
			if !err.IsNull() {
				return err
			}
			ci := ContainerInfo{InspectContainer:dock_inspect, ID:event.ID, Image:dock_inspect.Config.Image}
			tree_event.Trigger(&tree_event.Event{Name: tree_event.ON_DOCKER_CONTAINER_START, LocalVar: &ci})
		}
	case "die", "kill", "pause":
		{
			// Sending only Container ID if it stopped
			// Sometimes Docker API not giving all info about container after stopping it
			tree_event.Trigger(&tree_event.Event{Name: tree_event.ON_DOCKER_CONTAINER_STOP, LocalVar: event.ID})
		}
	case "pull", "tag":
		{
			var (
				inspect			*docker.Image
			)
			inspect, err.Err = DockerClient.InspectImage(event.ID)
			if !err.IsNull() {
				return err
			}
			im := ImageInfo{ID:inspect.ID, Name:event.ID, Inspect:inspect}
			tree_event.Trigger(&tree_event.Event{Name: tree_event.ON_DOCKER_IMAGE_CREATE, LocalVar: &im})
		}
	case "untag", "delete":
		{
			var (
				inspect			*docker.Image
			)
			inspect, err.Err = DockerClient.InspectImage(event.ID)
			if !err.IsNull() {
				return err
			}
			im := ImageInfo{ID:inspect.ID, Name:event.ID, Inspect:inspect}
			tree_event.Trigger(&tree_event.Event{Name: tree_event.ON_DOCKER_IMAGE_DELETE, LocalVar: &im})
		}
	}
	return err
}