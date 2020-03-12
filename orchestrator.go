/*

The main() function in this file is the entry point of the executable program for EDIRO.
This program performs the following tasks upon execution:
1. Starts the server and client functionality on the edge node that runs indefinitely awaiting inputs to be processed
2. Parse client requests and IoT resources uploads captured in a file.
3. Triggers the core modules of EDIRO as go routines

Author : Niket Agrawal

*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/niketagrawal/EDIRO/parser"
	"github.com/niketagrawal/EDIRO/resourcediscovery"
	"github.com/niketagrawal/EDIRO/resourcemanager"
	"github.com/niketagrawal/EDIRO/taskinitiator"
)

//IoTResources : A struct that contains array of IoT resources as stored in the input json file
type IoTResources struct {
	IoTResourcearray []resourcemanager.Newresource `json:"iotresources"`
}

//MonitorMem : This function monitors the run time memory usage by the program and prints out the statistics
//Source: https://scene-si.org/2018/08/06/basic-monitoring-of-go-apps-with-the-runtime-package/
func MonitorMem(duration int) {

	var m runtime.MemStats
	var interval = time.Duration(duration) * time.Second
	for {
		<-time.After(interval)

		// Read full mem stats
		runtime.ReadMemStats(&m)
		Numgoroutine := runtime.NumGoroutine()

		fmt.Printf("\tTotalAlloc : %v", m.TotalAlloc/1024/1024)
		fmt.Printf("\tSys : %v", m.Sys/1024/1024)
		fmt.Printf("\tHeapSys : %v", m.HeapSys/1024)
		fmt.Printf("\tStackInuse : %v", m.StackInuse/1024)
		fmt.Printf("\tHeapAlloc : %v", m.HeapAlloc/1024/1024)
		fmt.Printf("\tNextGC : %v", m.NextGC)
		fmt.Printf("\tNumGC : %v", m.NumGC)
		fmt.Printf("\tNumGoroutines : %v\n", Numgoroutine)

	}
}

/*
parseiotresources : This function parses IoT resources from the input json file in which the resources are stored in an
array of structures and writes to the new IoT resource arrival channel.
Input: unmarshalled struct converted from json, New IoT resource arrival channel
Output: Nil
*/
func parseiotresources(iotresources IoTResources, ch chan resourcemanager.Newresource) {
	for i := 0; i < len(iotresources.IoTResourcearray); i++ {
		fmt.Println("Resource: " + iotresources.IoTResourcearray[i].Resource)
		fmt.Println("NodeID: " + iotresources.IoTResourcearray[i].NodeID)
		var resource = resourcemanager.Newresource{Resource: iotresources.IoTResourcearray[i].Resource,
			NodeID: iotresources.IoTResourcearray[i].NodeID}
		ch <- resource
	}
}

/*
parseclientrequests : This function parses client requests from the input json file in which the requests
are stored as array of strings and writes to the new client request channel of type string
Input: unmarshalled struct converted from json, new client request channel
Output: Nil
*/
func parseclientrequests(clientrequests []string, ch chan string) {
	for i := 0; i < len(clientrequests); i++ {
		fmt.Println("Client Request: " + clientrequests[i])
		ch <- clientrequests[i]
		time.Sleep(3 * time.Second) //inter-arrival time between two consecutive client requests, use as desired
	}
}

func main() {

	go MonitorMem(1) //collect and print run time memory usage statistics every 1 second

	go resourcemanager.Init()

	<-resourcemanager.Done1 // waiting for the Init() goroutine to start the server in the background

	time.Sleep(4 * time.Second) // sufficient time for servers to setup first so that incoming client requests will be served surely

	var mux sync.Mutex

	chanparseroutput := make(chan parser.Parseroutput, 10)

	chandiscovery := make(chan resourcediscovery.Resourcediscoveryoutput, 10)

	/* chanNewIotResourceArrival - The input side of the resource manager, ie, facing the outside world. The information about
	arrival of new IoT resources is parsed by 'parseiotresources()', packaged into a struct and written to this channel
	*/
	chanNewIotResourceArrival := make(chan resourcemanager.Newresource, 10)

	/*chanNewIoTResourceUpdate - Corresponds to the output side of the resource manager which talks to other modules, ie, facing the other modules
	in the orchestrator. The information about new client requests ia parsed by 'parseclientrequests()' and written to this channel.
	The data from this channel is then fetched by 'Newresourceupdate()'
	*/
	chanNewIoTResourceUpdate := make(chan resourcemanager.Newresource, 10)

	//Channel to store the new client requests arriving at the system. Data from this channel is consumed by the parser.
	chanNewClientRequest := make(chan string, 10)

	IotResourcelist, err := os.Open("input.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened input.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer IotResourcelist.Close()
	byteValue, _ := ioutil.ReadAll(IotResourcelist)

	var iotresources IoTResources
	json.Unmarshal(byteValue, &iotresources)

	ClientRequestslist, errr := os.Open("clientrequest.json")
	if errr != nil {
		fmt.Println(errr)
	}
	fmt.Println("Successfully Opened clientrequests.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer ClientRequestslist.Close()
	byteValuee, _ := ioutil.ReadAll(ClientRequestslist)

	var clientrequests []string
	json.Unmarshal(byteValuee, &clientrequests)

	/*Starting all the gorouotines here at once. They will start processing the data as and when it arrives on the respective
	/channels they consume from */
	//go resourcediscovery.DetectDuplicateApp(chanparseroutput, chanduplicate)
	//go resourcediscovery.Discoverresource(chanduplicate, chandiscovery, &mux)
	go resourcediscovery.Discoverresource(chanparseroutput, chandiscovery, &mux)
	go taskinitiator.Createlaunchcommand(chandiscovery)
	go resourcemanager.Newresourceupdate(chanNewIotResourceArrival, &mux, chanNewIoTResourceUpdate)
	go parser.Parseinput(chanNewClientRequest, chanparseroutput)

	//Parse IoT resouces uploaded
	go parseiotresources(iotresources, chanNewIotResourceArrival)

	time.Sleep(2 * time.Second) //added to make sure all the resources are uploaded before taking in client requests

	//Parse client requests in parallel
	go parseclientrequests(clientrequests, chanNewClientRequest)

	<-resourcemanager.Done // to ensure we wait for server to shut down and only then the main() exists
}
