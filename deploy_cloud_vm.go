package main

import ( 
	//"fmt"
	//"log"
	"github.com/Tfindelkind/ntnx-golang-client-sdk"
	//"flag"
	)
	
const (
	NutanixHost string = "192.168.178.130"
	username 		   = "admin"
	password		   = "nutanix/4u"	
	VMname			   = "Test"	
	)	

func main() {	
	
	var n 		ntnxAPI.NTNXConnection
	var v 		ntnxAPI.VM_json_AHV

	
	n.NutanixHost = NutanixHost
	n.Username = username
	n.Password = password
	ntnxAPI.EncodeCredentials(&n)
	ntnxAPI.CreateHttpClient(&n)
	
	/*
	   1. Upload Image when file is specified and wait
	   2. Upload Cloud seed.iso to image 
	   2. Create VM and wait
	   3. Clone Image to Disk and wait
	   4. Attach seed.iso
	   5. Add network
	   6. start VM
	 */
	
	v.Config.Name = "createVM"
	v.Config.Description = "create VM by create_vm script hopefully with deploy script"
	v.Config.MemoryMb = 1
	v.Config.NumVcpus = 1
	v.Config.NumCoresPerVcpu = 1
	
	ntnxAPI.CreateVM_AHV(&n, &v)
}
