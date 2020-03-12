/*
This package implements the functionality to execute the workloads corresponding to the client requests and provide
dedicated service management feature by monitoring updates to the IoT resource while the workload is active.
To execut the workloads, Docker Swarm API to create and run a service are used. The service creation command is
constructed from the metadata collected by the edge nodes. Other commands are used to monitor the completion of the service
which is used to implement the resource monitoring feature.
It also measure the pipeline execution time which is the time spent in offloading a client's request.

Author : Niket Agrawal
*/

package taskinitiator

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/niketagrawal/EDIRO/library"
	"github.com/niketagrawal/EDIRO/resourcediscovery"
	"github.com/niketagrawal/EDIRO/resourcemanager"
)

var start time.Time

/*
launchtask: This function to handle the task of launch the containerized workload for each client request
*/
func launchtask(c resourcediscovery.Resourcediscoveryoutput) {
	isComplete := make(chan bool) //making the channel here as it closes in the child gorotuine track completion, so for
	//every nstance of this loop this channel will be created again

	chti := make(chan string, 10) //channel to carry the output of this function. The output is the service name
	//that is launched

	image := c.Applicationtolaunch
	servicename := c.Request
	targetnode := c.Locationtolaunch

	//fetch name of image and constraint of where to launch from channel and populate in the command below

	elapsed := time.Since(start)
	fmt.Println("pipeline execution time until execution of system command is: ", c.Request, elapsed)

	out, err := exec.Command("docker", "service", "create", "--name", servicename, "--restart-condition", "none", "--detach",
		"--constraint", targetnode, image).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Println("Command Successfully Executed")

	//find resoruce corresponding to this service
	resource := library.ApptoResource[image]

	go trackcompletion(servicename, isComplete)

	go resourcemanager.ResourceMonitor(chti, isComplete)

	// write to channel about the resource in use correspondig to this service
	chti <- resource

	output := string(out[:])
	fmt.Println(output)

	elapsed = time.Since(start)
	fmt.Println("pipeline execution time for request is: ", c.Request, elapsed)
}

/*Createlaunchcommand : performs the following tasks:
1. Constructs system commands to launch containers
2. Starts a go routine to track its completion
3. Starts a go routine to monitor updates to the resource in use by this service. This go routine lasts
until the previous go routine runs
4. Maintains a mapping of application currently running with the corresponding request to aid in resource
monitoring concurrently
Input : a structure encapsulating the following details:
- Application/dockerfile to launch as container
- target node in the cluster where this containerized application will be executed
- client request which forms the name of the launched service
Output: Nil
*/
func Createlaunchcommand(ch chan resourcediscovery.Resourcediscoveryoutput) {
	start = time.Now()
	for {
		c := <-ch

		go launchtask(c) //spawning a new goroutine to handle each client request to avoid sequential
		//processing and other requests waiting in the queue behind the current request being processed

	}

}

/* trackcompletion : This function tracks completion of a service
Input : service launched
Output : done with service name written to a channel, check if we have channel for dedicated service then passing service
name to channel isn't needed , just a done is fine
*/
func trackcompletion(servicename string, isComplete chan bool) {
	fmt.Println("tracking completion of : ", servicename)
	for {
		out, err := exec.Command("docker", "service", "ps", servicename).Output()
		if err != nil {
			fmt.Printf("%s", err)
		}
		output := string(out[:])
		status := strings.Contains(output, "Complete")
		if status {
			fmt.Println("application completed, stopping resource monitoring by closing channel", servicename)
			close(isComplete) //closing channel to signal completion of application
			break
		} else {
			continue
		}

	}
	fmt.Println("BROKE out of for loop to signal completion of :", servicename)
}
