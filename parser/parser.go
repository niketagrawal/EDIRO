/*

This package implementas the parser module.
Parser: Parses the requests and fetches information about the corresponding application packages that needs to be deployed
on the edge nodes using the configuration library.

Author : Niket Agrawal

*/

package parser

import (
	"github.com/niketagrawal/EDIRO/library"
)

/*Parseroutput : The output of the parser is modelled as a structure that contains the client request, the
corresponding application package and the associated IoT resource.
*/
type Parseroutput struct {
	Request, Application, Resource string
}

//Parseinput : This function parses the client reqests, performs a map look up and renders the application and the
//associated IoT resource corresponding to this client request
func Parseinput(chIn chan string, chanparseroutput chan Parseroutput) {
	for {
		request := <-chIn
		requiredapp := library.RequesttoApp[request]
		requiredresource := library.ApptoResource[requiredapp]
		var output Parseroutput
		output.Application = requiredapp
		output.Resource = requiredresource
		output.Request = request
		chanparseroutput <- output
	}

}
