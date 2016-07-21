package main

import (
	// "log"
	//"archive/tar"
	//"strconv"
	//"errors"
	"fmt"
	//"net"
	//"net/http"
	//"io/ioutil"
	//"io"
	//"os"
	//"time"
	//"path/filepath"
	
	// rename because duplicated package name
	//gssh "golang.org/x/crypto/ssh"
	
	"github.com/Tfindelkind/ntnx-golang-client-sdk"
)

const (
	defaultUser   			= "admin"
	defaultPass   			= "nutanix/4u"
	defaultHost				= "192.168.178.130"	
)


func main() {
	
	var n 		ntnxAPI.NTNXConnection
		
	fmt.Printf("Setup Nutanix REST connection...")
	
	n.NutanixHost = defaultHost
	n.Username = defaultUser
	n.Password = defaultPass
	ntnxAPI.EncodeCredentials(&n)
	ntnxAPI.CreateHttpClient(&n)
		
    fmt.Println(ntnxAPI.GetVMsbyContainer(&n,"ISO"))
    
    
	
}




