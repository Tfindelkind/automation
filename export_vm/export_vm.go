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
	"io"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	ntnxAPI "github.com/Tfindelkind/acropolis-sdk-go"
	"golang.org/x/crypto/ssh"
)

const appVersion = "0.9 beta"

var (
	host        *string
	username    *string
	password    *string
	sshusername *string
	sshpassword *string
	vmName      *string
	container   *string
	esxhost     *string
	esxusername *string
	esxpassword *string
	diskformat  *string
	debug       *bool
	overwrite   *bool
	help        *bool
	version     *bool
)

func init() {
	host = flag.String("host", "", "a string")
	username = flag.String("username", "", "a string")
	password = flag.String("password", "", "a string")
	sshusername = flag.String("sshusername", "", "a string")
	sshpassword = flag.String("sshpassword", "", "a string")
	vmName = flag.String("vm-name", "", "a string")
	container = flag.String("container", "", "a string")
	esxhost = flag.String("esxhost", "", "a string")
	esxusername = flag.String("esxusername", "", "a string")
	esxpassword = flag.String("esxpassword", "", "a string")
	diskformat = flag.String("diskformat", "", "a string")
	debug = flag.Bool("debug", false, "a bool")
	overwrite = flag.Bool("overwrite", false, "a bool")
	help = flag.Bool("help", false, "a bool")
	version = flag.Bool("version", false, "a bool")
}

func printHelp() {

	fmt.Println("Usage: export_vm [OPTIONS]")
	fmt.Println("export_vm [ --help | --version ]")
	fmt.Println("")
	fmt.Println("FOR NUTANIX AHV ONLY- exports an AHV VM")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("")
	fmt.Println("--host             Specify CVM host or Cluster IP")
	fmt.Println("--username         Specify username for connect to host")
	fmt.Println("--password         Specify password for user")
	fmt.Println("--sshusername      Specify ssh username for connect to host")
	fmt.Println("--sshpassword      Specify ssh password for ssh username")
	fmt.Println("--vm-name          Specify Virtual Machine name which will exported")
	fmt.Println("--container        Specify the container where vm will be exported to")
	fmt.Println("--diskformat       Specify the diskformat vmdk or qcow2")
	fmt.Println("--debug            Enables debug mode")
	fmt.Println("--overwrite		    Overwrites target VM/Images (delete and creates new one)")
	fmt.Println("--help             List this help")
	fmt.Println("--version          Show the deploy_cloud_vm version")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("")
	fmt.Println("export_vm --host=NTNX-CVM --username=admin --password=nutanix/4u --vm-name=MyVM --container=ISO")
	fmt.Println("")
}

func evaluateFlags() (ntnxAPI.NTNXConnection, ntnxAPI.VMJSONAHV) {

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

	//sshusername
	if *sshusername == "" {
		log.Warn("option '--sshusername=' is not set  Default: nutanix is used")
		*sshusername = "nutanix"
	}

	//sshpassword
	if *sshpassword == "" {
		log.Warn("option '--sshusername=' is not set  Default: nutanix/4u is used")
		*sshpassword = "nutanix/4u"
	}

	//vm-name
	if *vmName == "" {
		log.Warn("mandatory option '--vm-name=' is not set")
		os.Exit(0)
	}

	//diskformat
	if *diskformat == "" {
		log.Warn("option '--diskformat=' is not set  Default: vmdk is used")
		*diskformat = "vmdk"
	} else {
		if *diskformat != "qcow2" && *diskformat != "vmdk" {
			log.Fatal("diskformat: " + *diskformat + " is unknown.")
		}
	}

	//container
	if *container == "" {
		log.Warn("mandatory option '--container=' is not set")
		os.Exit(0)
	}

	var vm ntnxAPI.VMJSONAHV
	vm.Config.Name = *vmName

	var n ntnxAPI.NTNXConnection

	n.NutanixHost = *host
	n.Username = *username
	n.Password = *password

	ntnxAPI.EncodeCredentials(&n)
	ntnxAPI.CreateHTTPClient(&n)

	ntnxAPI.NutanixCheckCredentials(&n)

	// If container is not found exit
	if *container != "" {
		_, err := ntnxAPI.GetContainerUUIDbyName(&n, *container)
		if err != nil {
			os.Exit(1)
		}
	}

	return n, vm

}

func sshExec(host string, user string, password string, exec string) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}

	connection, err := ssh.Dial("tcp", host+":22", sshConfig)
	if err != nil {
		log.Warn("Failed to dial: %s")
		os.Exit(1)
	}

	session, err := connection.NewSession()
	if err != nil {
		log.Warn("Failed to create session: %s")
		os.Exit(1)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Warn("Unable to setup stdin for session: %v")
		os.Exit(1)
	}
	go io.Copy(stdin, os.Stdin)

	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Warn("Unable to setup stdout for session: %v")
		os.Exit(1)
	}
	go io.Copy(os.Stdout, stdout)

	stderr, err := session.StderrPipe()
	if err != nil {
		log.Warn("Unable to setup stderr for session: %v")
		os.Exit(1)
	}
	go io.Copy(os.Stderr, stderr)

	session.Run(exec)
}

func main() {

	flag.Usage = printHelp
	flag.Parse()

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	var n ntnxAPI.NTNXConnection
	var v ntnxAPI.VMJSONAHV
	//	var d ntnxAPI.VDiskJSONREST
	//var net ntnxAPI.NetworkREST
	//	var taskUUID ntnxAPI.TaskUUID
	var existV bool

	n, v = evaluateFlags()

	/*
			   Short description what will be done

			   1. Check if VM exist - Check if powered on
		     2. analyze VM and get config with all disks, vmnic
		     3. export each vmdk to container
		     4. Whitelist ESX and mount
		     3. show download path

		     4. optional show
	*/

	//check if source VM exists
	existV, _ = ntnxAPI.VMExist(&n, v.Config.Name)

	if existV {

		v, _ = ntnxAPI.GetVMbyName(&n, &v)

		// check if Source VM is running
		state := ntnxAPI.GetVMState(&n, &v)

		if state != "off" {
			log.Warn("VM " + v.Config.Name + " is not powered off")
			os.Exit(0)
		}

		var i int

		// export each vDisk to NFS store
		for _, elem := range v.Config.VMDisks {
			if !elem.IsCdrom {

				TargetContainerName := *container

				SourceContainerName, _ := ntnxAPI.GetContainerNamebyUUID(&n, elem.ContainerUUID)

				i++
				execString := "/usr/local/nutanix/bin/qemu-img convert nfs://127.0.0.1/" + SourceContainerName + "/.acropolis/vmdisk/" + elem.VMDiskUUID + " -O " + *diskformat + " nfs://127.0.0.1/" + TargetContainerName + "/" + v.Config.Name + "-" + strconv.Itoa(i) + "." + *diskformat

				log.Info("Converting disk: " + strconv.Itoa(i) + " from VM: " + v.Config.Name + " to disk /" + TargetContainerName + "/" + v.Config.Name + "-" + strconv.Itoa(i) + "." + *diskformat)
				sshExec(*host, *sshusername, *sshpassword, execString)

			}
		}
	}
}
