/*

This packagae implements the resource discovery module of EDIRO.
It performs the following tasks:
1. It discovers the location of the IoT resource on the edge cluster needed to execute an application
to satisfy a client request. This determines the offloading location for the workload.
2. It detects if the client request can be served by an ongiong workload on the edge cluster.

Author : Niket Agrawal

*/

package resourcediscovery

import (
	"fmt"
	"sync"

	"github.com/niketagrawal/EDIRO/parser"
	"github.com/niketagrawal/EDIRO/resourcemanager"
)

var targetnode int

type Resourcediscoveryoutput struct {
	Request, Applicationtolaunch, Locationtolaunch string
}

/*
DiscoverresourcesubGoroutine : function to which the task of performing resource discovery is delegated,
runs as a go routine
*/
func DiscoverresourcesubGoroutine(s parser.Parseroutput, chandiscov chan Resourcediscoveryoutput,
	m *sync.Mutex) {
	var targetnode string
out:
	for key := range resourcemanager.Resourcetable {
		for i := range resourcemanager.Resourcetable[key] {
			if s.Resource == resourcemanager.Resourcetable[key][i] {
				targetnode = key
				m.Lock()
				resourcemanager.Resourcetable[key][i] = "used" //adding a label to mark the IoT resource as used
				//and avoid it being detected by the resource monitoring algorithm. Acquired Lock to ensure an
				//atomic operation.
				m.Unlock()
				fmt.Println("targetnode is :", targetnode)
				break out
			}
		}
	}

	var out Resourcediscoveryoutput
	out.Applicationtolaunch = s.Application
	out.Locationtolaunch = targetnode
	out.Request = s.Request
	fmt.Println("Application and target node to launch are:", out)

	chandiscov <- out

}

//Discoverresource : It determines the presence and location of the IoT resource needed by an application.
//Input: Receives a signal from detect duplicate function whether a fresh application needs to be launched or not
//Output: provides the location to luanch a particular application. Request and application to launch are supplied as complimentary
func Discoverresource(chanpo chan parser.Parseroutput, chandiscov chan Resourcediscoveryoutput, m *sync.Mutex) {
	for {
		s := <-chanpo //acts on output from detect duplicate function

		go DiscoverresourcesubGoroutine(s, chandiscov, m)

	}

}
