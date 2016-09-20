/* Copyright (c) 2016 Thomas Findelkind
#
# This program is free software: you can redistribute it and/or modify it under
# the terms of the GNU General Public License as published by the Free Software
# Foundation, either version 3 of the License, or (at your option) any later
# version.
#
# This program is distributed in the hope that it will be useful, but WITHOUT
# ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
# FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
# details.
#
# You should have received a copy of the GNU General Public License along with
# this program.  If not, see <http://www.gnu.org/licenses/>.
#
# MORE ABOUT THIS SCRIPT AVAILABLE IN THE README AND AT:
#
# http://tfindelkind.com
#
# ----------------------------------------------------------------------------
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	ntnxAPI "github.com/Tfindelkind/acropolis-sdk-go"
)

const appVersion = "0.9 beta"

var (
	host       *string
	username   *string
	password   *string
	container  *string
	mountpoint *string
	whitelist  *string
	debug      *bool
	help       *bool
	version    *bool
)

func init() {
	host = flag.String("host", "", "a string")
	username = flag.String("username", "", "a string")
	password = flag.String("password", "", "a string")
	container = flag.String("container", "", "a string")
	mountpoint = flag.String("mountpoint", "", "a string")
	whitelist = flag.String("whitelist", "", "a string")
	debug = flag.Bool("debug", false, "a bool")
	help = flag.Bool("help", false, "a bool")
	version = flag.Bool("version", false, "a bool")
}

func printHelp() {

	fmt.Println("Usage: mount_nfs [OPTIONS]")
	fmt.Println("mount_nfs [ --help | --version ]")
	fmt.Println("")
	fmt.Println("FOR NUTANIX AHV ONLY- exports an AHV VM")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("")
	fmt.Println("--host             Specify CVM host or Cluster IP")
	fmt.Println("--username         Specify username for connect to host")
	fmt.Println("--password         Specify password for user")
	fmt.Println("--container        Specify the container to mount - Default mount all")
	fmt.Println("--mountpoint       (Optional) the mount point like ´/mount/´ WITH tailing /")
	fmt.Println("--whitelist		    (Optional) nnn.nnn.nnn.nnn/xxx.xxx.xxx.xxx")
	fmt.Println("           		    where nnn is the IP address, and xxx is the subnet mask.")
	fmt.Println("--debug            Enables debug mode")
	fmt.Println("--help             List this help")
	fmt.Println("--version          Show the deploy_cloud_vm version")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("")
	fmt.Println("mount_nfs --host=NTNX-CVM --username=admin --password=nutanix/4u")
	fmt.Println("")
}

func evaluateFlags() ntnxAPI.NTNXConnection {

	//help
	if *help {
		printHelp()
		os.Exit(0)
	}

	//version
	if *version {
		fmt.Println("Version: " + appVersion)
		os.Exit(0)
	}

	//debug
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	//host
	if *host == "" {
		log.Warn("mandatory option '--host=' is not set")
		os.Exit(0)
	}

	//username
	if *username == "" {
		log.Warn("option '--username=' is not set  Default: admin is used")
		*username = "admin"
	}

	//password
	if *password == "" {
		log.Warn("option '--password=' is not set  Default: nutanix/4u is used")
		*password = "nutanix/4u"
	}

	//container
	if *container == "" {
		log.Warn("option '--container=' is not set  Default: mounting all")
		*container = "MOUNT-ALL"
	}

	//mountpoint
	if *mountpoint == "" {
		log.Warn("option '--mountpoint=' is not set  Default: /mnt/<containername>")
	}

	var n ntnxAPI.NTNXConnection

	n.NutanixHost = *host
	n.Username = *username
	n.Password = *password

	ntnxAPI.EncodeCredentials(&n)
	ntnxAPI.CreateHTTPClient(&n)

	ntnxAPI.NutanixCheckCredentials(&n)

	return n

}

func mkDIR(path string) {

	_, err := exec.Command("/bin/bash", "-c", "sudo mkdir -p "+path).Output()
	if err != nil {
		log.Error("Could not create mountpoint: " + path)
	}
}

func mount(hostname string, share string, path string) {
	fmt.Println("/bin/bash -c sudo mount -t nfs " + hostname + ":/" + share + " " + path)
	_, err := exec.Command("/bin/bash", "-c", "sudo mount -t nfs "+hostname+":/"+share+" "+path).Output()
	if err != nil {
		log.Error("Could not mount share: " + share + " from host: " + hostname + " to path: " + path)
	}
}

func main() {

	flag.Usage = printHelp
	flag.Parse()

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	var n ntnxAPI.NTNXConnection

	n = evaluateFlags()

	//ntnxAPI.AddWhiteList(&n, "10.10.11.0/255.255.255.0")

	if *container != "MOUNT-ALL" {
		_, err := ntnxAPI.GetContainerIDbyName(&n, *container)
		if err != nil {
			os.Exit(1)
		}

		if *mountpoint != "" {
			mkDIR(*mountpoint + *container)
			mount(*host, *container, *mountpoint+*container)

		} else {
			mkDIR("/mnt/" + *container)
			mount(*host, *container, "/mnt/"+*container)

		}
		os.Exit(0)
	}

	list, _ := ntnxAPI.GetContainerNames(&n)
	for _, elem := range list {

		mkDIR("/mnt/" + elem)
		mount(*host, elem, "/mnt/"+elem)

	}

}
