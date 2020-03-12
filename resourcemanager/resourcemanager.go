/*
This packagae implements the Resource management module of EDIRO. It consists of the following sub modules and performs the
following tasks:
1. Inter edge communication
2. Spreading of IoT resource availability information when a new IoT resource is uploaded on this edge node
3. Handling of similar updates from other edge nodes
4. Monitoring of IoT resource for a running workload
5. Measure the time takn to spread the metadata about an IoT resource to other edge nodes.

Author: Niket Agrawal

Part of GRPC client server code is sourced from : https://github.com/grpc/grpc-go/tree/master/examples/helloworld
*/

package resourcemanager

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/niketagrawal/EDIRO/protobufferfile"

	"google.golang.org/grpc"
)

//Resourcetable : Declaring the map to store information about IoT Resource availabiity on each edge node
//in the cluster
var Resourcetable map[string][]string

//Specifying the litsening address of this edge node on which it listens for messages from other edge nodes
const (
	port = "1.1.1.1:1"
)

type server struct{}

//Done channel to signal main() when the go routine finishes.
var Done = make(chan bool)
var Done1 = make(chan bool)

var mux sync.Mutex

//Newresource : struct to hold the data format in which the newresourceupdate function will pack data in and send to
//broadcasting go routine on a channel
type Newresource struct {
	Resource, NodeID string
}

func (s *server) ResourceTableUpdate(ctx context.Context, in *pb.TableUpdate) (*pb.TableUpdateACK, error) {
	log.Printf("Received: %v %v", in.Resource, in.ID)
	Updatetableafterhearing(in.Resource, in.ID, &mux)
	return &pb.TableUpdateACK{Ack: "tableupdateACK" + in.Resource}, nil
}

/*
Listenforupdates : This function starts the listening server on the edge node to listen for updates from
other edge nodes about IoT resource availability
Source: https://github.com/grpc/grpc-go/tree/master/examples/helloworld
*/
func Listenforupdates() {
	fmt.Println("launching grpcserver for listening to updates")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterFrontendServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	Done <- true //signalling done here but this line gets hit only when we close the server
}

/*
Reachabililty details of all other edge nodes in the cluster in terms of their IP address and
listening ports is specified here.
*/
var address = [1]string{"2.2.2.2:2"}

/*
Broadcast : This function broadcasts the information about upload of a new IoT resource on this edge
node to all other edge nodes
Source: https://github.com/grpc/grpc-go/tree/master/examples/helloworld
*/
func Broadcast(ch chan Newresource, measurechannel chan bool) {
	input := <-ch

	var counter int
	//loop to send on all other edge nodes
	for i := 0; i < 1; i++ {
		// Establish a connection to the server.
		conn, err := grpc.Dial(address[i], grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewFrontendClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.ResourceTableUpdate(ctx, &pb.TableUpdate{Resource: input.Resource, ID: input.NodeID})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s", r.Ack)
		// update a global counter after every ACK recceived and when it reaches 'n-1' (n = no. of nodes in cluster)
		// update on the channel so that measure function stops the timer
		counter++
		if counter == 1 {
			measurechannel <- true
		}

	}

}

/*
Init function is called from orchestartor only once when orchestrator starts. This initializes the
IoT resource catalog or the local system state and starts a listener for receiving updates from other nodes
*/
func Init() {

	Resourcetable = map[string][]string{}

	//start listener in background
	go Listenforupdates()
	Done1 <- true
}

/*
Newresourceupdate : Handles the IoT resources offloaded on this edge node, updates the local state and broadcasts
this update to other edge nodes in the cluster.
Input: arrival of message on the channel dedicated for new IoT resources offloaded
Output: Nil
*/
func Newresourceupdate(chIn chan Newresource, m *sync.Mutex, chOut chan Newresource) {

	for {
		NewIoTResourceUpload := <-chIn

		measurechannel := make(chan bool) //channel to indicate about ACK reception on resource updation from
		//other nodes

		go MeasureTime(measurechannel, NewIoTResourceUpload.Resource)
		m.Lock()
		fmt.Println("Lock acquired by Newresourceupdate")
		fmt.Println("Newresourceupdate: Received iot resource and nodeID to append are:", NewIoTResourceUpload.Resource, NewIoTResourceUpload.NodeID)
		fmt.Println("Newresourceupdate: Map before appending on node 1 is : ", Resourcetable)
		fmt.Println("Newresourceupdate: Received IoT resource to append is :", NewIoTResourceUpload.Resource)

		res := append(Resourcetable[NewIoTResourceUpload.NodeID], NewIoTResourceUpload.Resource)
		Resourcetable[NewIoTResourceUpload.NodeID] = res
		m.Unlock()
		fmt.Println("Lock released by Newresourceupdate")
		fmt.Println("Newresourceupdate: slice after appending is :", res)
		fmt.Println("Newresourceupdate: map after appending is : ", Resourcetable)

		//broadcast this update
		var output Newresource
		output.Resource = NewIoTResourceUpload.Resource
		output.NodeID = NewIoTResourceUpload.NodeID
		chOut <- output
		go Broadcast(chOut, measurechannel)
	}

}

/*
MeasureTime : Measures time from the time of resource upload on a node to its update on all other edge ndoes.
Input: a boolean channel to indicate start and stop of timer, The associated iot resource for which the
spreading time is being measured
Output: Nil
*/
func MeasureTime(measurechannel chan bool, iotresource string) {
	start := time.Now()
	<-measurechannel
	elapsed := time.Since(start)
	fmt.Println("time elapsed in spreading resource is:", iotresource, elapsed)
}

/*
Updatetableafterhearing : Updates the resource catalog upon hearing new resource availability updates from
other nodes. This function is called by the server on this node upon reception of new resource updates.
Input: Received resource name and associated edge node ID
Output: Nil.
*/
func Updatetableafterhearing(input string, ID string, m *sync.Mutex) {
	m.Lock()
	fmt.Println("Updatetableafterhearing: Received resource and nodeID are: ", input, ID)
	fmt.Println("Updatetableafterhearing: Map before appending is:", Resourcetable)
	res := append(Resourcetable[ID], input)
	Resourcetable[ID] = res
	fmt.Println("Updatetableafterhearing: slice after appending is :", res)
	fmt.Println("Updatetableafterhearing: table after receiving update is : ", Resourcetable)
	m.Unlock()
}

/*
ResourceMonitor : It monitors the availability of a new version of an IoT resource that is currently in use
and comunicates its arrival to task initiator to take necessary actions.
Input: the resource to monitor as provided by task initiator module, channel of type bool which is closed
when application completes its execution signalling stopping of resource monitoring
Output: Nil (currently, only a statement is printed on the console to signal the new version of resource found)
*/
func ResourceMonitor(resourceToMonitor chan string, isComplete chan bool) {
	resourcetofind := <-resourceToMonitor
	for {
		select {
		case <-isComplete:
			fmt.Println("channel closed: application has terminated, no more resource monitoring required")
			return

		default:
			for key := range Resourcetable {
				for i := range Resourcetable[key] {
					if resourcetofind == Resourcetable[key][i] {
						fmt.Println("New version found of resource : ", resourcetofind)
						return
					}
				}
			}
		}
	}
}
