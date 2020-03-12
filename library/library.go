/*
This package implements the Library module of EDIRO
The library module maintains the information about the application packages corresponding to a particular
request in form of key-value pairs that we assume are made available by a knowledgeable party.

Author : Niket Agrawal

*/

package library

/*
  RequesttoApp : Maps incoming client request to the corresponding application that needs to be deployed to
   fullfil that request. Specify the name of the desired application image below
*/
var RequesttoApp = map[string]string{
	"client_request_1": "application_image_1",
	"client_request_2": "application_image_2",
	"client_request_3": "application_image_3",
}

/*
ApptoResource : Maps the application (dockerfile) to be run with the IoT resource (IoT data input) it requires.
The IoT resource is of the type string and points to the loation of that resource on that specific node.
*/
var ApptoResource = map[string]string{
	"application_image_1": "IoT_resource_1",
	"application_image_2": "IoT_resource_2",
	"application_image_3": "IoT_resource_3",
}
