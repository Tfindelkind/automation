package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	// "log"
	//"archive/tar"
	//"strconv"
	//"errors"
	"os/exec"
	//"net"
	//"net/http"
	//"io/ioutil"
	//"io"
	//"time"
	//"path"
	
	// rename because duplicated package name
	//gssh "golang.org/x/crypto/ssh"
	
	//"github.com/Tfindelkind/ntnx-golang-client-sdk"
	
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	//github.com/vmware/govmomi/property"
	//"github.com/vmware/govmomi/list"
	//"github.com/vmware/govmomi/vim25"
	//"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	//"github.com/vmware/govmomi/object"
	"golang.org/x/net/context"
	"github.com/vmware/govmomi/govc/flags"
)


// GetEnvString returns string from environment variable.
func GetEnvString(v string, def string) string {
	r := os.Getenv(v)
	if r == "" {
		return def
	}

	return r
}

// GetEnvBool returns boolean from environment variable.
func GetEnvBool(v string, def bool) bool {
	r := os.Getenv(v)
	if r == "" {
		return def
	}

	switch strings.ToLower(r[0:1]) {
	case "t", "y", "1":
		return true
	}

	return false
}

func printCommand(cmd *exec.Cmd) {
  fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printError(err error) {
  if err != nil {
    os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
  }
}

func printOutput(outs []byte) {
  if len(outs) > 0 {
    fmt.Printf("==> Output: %s\n", string(outs))
  }
}


const (
	ntnxUserName  			= "admin"
	ntnxPassword  			= "nutanix/4u"
	ntnxHost				= "192.168.178.130"	
)


const (
	vmwareEnvURL      = "GOVMOMI_URL"
	vmwareEnvUserName = "GOVMOMI_USERNAME"
    vmwareEnvPassword = "GOVMOMI_PASSWORD"
	vmwareEnvInsecure = "GOVMOMI_INSECURE"
)

const (
	vmwareURL      = "https://192.168.178.80/sdk"
	vmwareUserName = "root"
	vmwarePassword = "nutanix/4u"
	vmwareInsecure = true
)

var urlDescription = fmt.Sprintf("ESX or vCenter URL [%s]", vmwareEnvURL)
var urlFlag = flag.String("url", GetEnvString(vmwareEnvURL, vmwareURL), urlDescription)

var insecureDescription = fmt.Sprintf("Don't verify the server's certificate chain [%s]", vmwareEnvInsecure)
var insecureFlag = flag.Bool("insecure", GetEnvBool(vmwareEnvInsecure, vmwareInsecure), insecureDescription)

func processOverride(u *url.URL) {
	envUsername := GetEnvString(vmwareEnvUserName,vmwareUserName)
	envPassword := GetEnvString(vmwareEnvPassword,vmwarePassword)

	// Override username if provided
	if envUsername != "" {
		var password string
		var ok bool

		if u.User != nil {
			password, ok = u.User.Password()
		}

		if ok {
			u.User = url.UserPassword(envUsername, password)
		} else {
			u.User = url.User(envUsername)
		}
	}

	// Override password if provided
	if envPassword != "" {
		var username string

		if u.User != nil {
			username = u.User.Username()
		}

		u.User = url.UserPassword(username, envPassword)
	}
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}

type change struct {
	*flags.DatacenterFlag

	types.ClusterConfigSpecEx
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()

	// Parse URL from string
	u, err := url.Parse(*urlFlag)
	if err != nil {
		exit(err)
	}

	// Override username and/or password as required
	processOverride(u)

	// Connect and log in to ESX or vCenter
	c, err := govmomi.NewClient(ctx, u, *insecureFlag)
	if err != nil {
		exit(err)
	}

	f := find.NewFinder(c.Client, true)
	
		// Find one and only datacenter
	dc, err := f.DefaultDatacenter(ctx)
	if err != nil {
		exit(err)
	}

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)


 /* Start */
 
	/*// Find datastores in datacenter
	dss, err := f.DatastoreList(ctx, "*")
	if err != nil {
		exit(err)
	}*/
	
	
	/*clusters, err := f.ClusterComputeResourceList(ctx, "*")
			
	
		for _, cluster := range clusters {
				
			fmt.Println(&cmd.ClusterConfigSpecEx)
		}*/
	

	
	
	/*var n 		ntnxAPI.NTNXConnection
		
	fmt.Printf("Setup Nutanix REST connection...")
	
	n.NutanixHost = ntnxHost
	n.Username = ntnxUserName
	n.Password = ntnxPassword
	ntnxAPI.EncodeCredentials(&n)
	ntnxAPI.CreateHttpClient(&n)
		
    fmt.Println(ntnxAPI.GetVMsbyContainer(&n,"ISO"))
    
	cmd := exec.Command("bash","-c","govc about -k")
	
	// Combine stdout and stderr
	printCommand(cmd)
	output, err := cmd.CombinedOutput()
	printError(err)
	printOutput(output) // => go version go1.3 darwin/amd64*/
}






