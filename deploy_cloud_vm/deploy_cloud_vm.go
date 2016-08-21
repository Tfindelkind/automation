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
	log "github.com/Sirupsen/logrus"
	"github.com/Tfindelkind/ntnx-golang-client-sdk"

	"fmt"
	"os"
	//"time"
	"flag"
)

const AppVersion = "0.9 beta"

var (
	host            *string
	username        *string
	password        *string
	vm_name         *string
	image_name      *string
	seed_name       *string
	image_file      *string
	seed_file       *string
	vlan            *string
	container       *string
	debug           *bool
	help            *bool
	version         *bool
)

func init() {
	host = flag.String("host", "192.168.178.130", "a string")
	username = flag.String("username", "admin", "a string")
	password = flag.String("password", "nutanix/4u", "a string")
	vm_name = flag.String("vm-name", "NTNX-AVM", "a string")
	image_name = flag.String("image-name", "Centos7-1606", "a string")
	seed_name = flag.String("seed-name", "CloudInitSeed", "a string")
	image_file = flag.String("image-file", "CentOS-7-x86_64-GenericCloud-1606.qcow2", "a string")
	seed_file = flag.String("seed-file", "seed.iso", "a string")
	vlan = flag.String("vlan", "VLAN0", "a string")
	container = flag.String("container", "ISO", "a string")	
	debug = flag.Bool("debug", false, "a bool")
	help = flag.Bool("help", false, "a bool")
	version = flag.Bool("version", false, "a bool")
}

func printHelp() {

	fmt.Println("Usage: deploy_cloud_vm [OPTIONS]")
	fmt.Println("deploy_cloud_vm [ --help | --version ]")
	fmt.Println("")
	fmt.Println("Upload and deploy a cloud image with a CD seed")
	fmt.Println("Example seed.iso at https://github.com/Tfindelkind/DCI")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("")
	fmt.Println("--host             Specify CVM host or Cluster IP")
	fmt.Println("--username         Specify username for connect to host")
	fmt.Println("--password         Specify password for user")
	fmt.Println("--vm-name          Specify Virtual Machine name which will be created")
	fmt.Println("--image-name       Specify the name of the Cloud-Image in Image Service")
	fmt.Println("--image-file       Speficy the file name of the Cloud-Image")
	fmt.Println("--seed-name        Specify the name of the Seed.ISO in Image Service")
	fmt.Println("--seed-file        Speficy the file name of the Seed.ISO")
	fmt.Println("--vlan             Specify the VLAN to which the VM will be connected")
	fmt.Println("--container        Specify the container where images/vm will be stored")
	fmt.Println("--debug            Enables debug mode")
	fmt.Println("--help             List this help")
	fmt.Println("--version          Show the deploy_cloud_vm version")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("")
	fmt.Println("deploy_cloud_vm --host=NTNX-CVM --username=admin --password=nutanix/4u --vm-name=NTNX-AVM --image-name=Centos7-1606 --image-container=ISO --vm-container=prod vlan=VLAN0")
	fmt.Println("")
}

func main() {

	flag.Usage = printHelp
	flag.Parse()
		

	if *help {
		printHelp()
		os.Exit(0)
	}

	if *version {
		fmt.Println("Version: " + AppVersion)
		os.Exit(0)
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	var n ntnxAPI.NTNXConnection
	var v ntnxAPI.VM_json_AHV
	var nic1 ntnxAPI.Network_REST
	var im ntnxAPI.Image_json_AHV
	var seed ntnxAPI.Image_json_AHV
	var taskUUID ntnxAPI.TaskUUID

	n.NutanixHost = *host
	n.Username = *username
	n.Password = *password
	im.Name = *image_name
	im.Annotation = "deployed with deploy_cloud_vm"
	im.ImageType = "DISK_IMAGE"
	seed.Name = *seed_name
	seed.Annotation = "deployed with deploy_cloud_vm"
	seed.ImageType = "ISO_IMAGE"
	v.Config.Name = *vm_name
	v.Config.Description = "deployed with deploy_cloud_vm"
	v.Config.MemoryMb = 4096
	v.Config.NumVcpus = 1
	v.Config.NumCoresPerVcpu = 1

	ntnxAPI.EncodeCredentials(&n)
	ntnxAPI.CreateHttpClient(&n)

	/*
	   Short description what will be done

	   1. Upload Image when file is specified and wait
	   2. Upload Cloud seed.iso to image
	   2. Create VM and wait
	   3. Clone Image to Disk and wait
	   4. Attach seed.iso
	   5. Add network
	   6. start VM
	*/

	/*To-DO:
	
	  1. Inplement progress bar while uploading- (concurreny and get progress from task)


	*/

	// upload cloud image to image service
	if ntnxAPI.ImageExistbyName(&n, &im) {
		log.Warn("Image " + im.Name + " already exists")
		// get existing image ID		
	} else {
		taskUUID, _ = ntnxAPI.CreateImageObject(&n, &im)

		task, err := ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
		if err != nil {
			log.Fatal("Task does not exist")
		}

		im.UUID = ntnxAPI.GetImageUUIDbyTask(&n, &task)

		_, statusCode := ntnxAPI.PutFileToImage(&n, ntnxAPI.NutanixAHVurl(&n), "images/"+im.UUID+"/upload", *image_file, *container)

		if statusCode != 200 {
			log.Error("Image upload failed")
			os.Exit(1)
		}					
	}

	// upload seed.iso to image service
	if ntnxAPI.ImageExistbyName(&n, &seed) {
		log.Warn("Image " + seed.Name + " already exists")
	} else {
		taskUUID, _ = ntnxAPI.CreateImageObject(&n, &seed)

		task, err := ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
		if err != nil {
			log.Fatal("Task does not exist")
		}

		seed.UUID = ntnxAPI.GetImageUUIDbyTask(&n, &task)

		_, statusCode := ntnxAPI.PutFileToImage(&n, ntnxAPI.NutanixAHVurl(&n), "images/"+seed.UUID+"/upload", *seed_file, *container)

		if statusCode != 200 {
			log.Error("Image upload failed")
			os.Exit(1)
		}
	}
	
	// make sure cloud image is active and get all infos when active
	log.Info("Wait that the cloud image is activated...")
	ImageActive, _ := ntnxAPI.WaitUntilImageIsActive(&n,&im)
	if ( !ImageActive ) {
		log.Fatal("Cloud Image is not active")
		os.Exit(1)
	}
	im , _ = ntnxAPI.GetImagebyName(&n, im.Name)
	
	// make sure seed image is active and get all infos when active	
	log.Info("Wait that the seed image is activated...")
	ImageActive, _ = ntnxAPI.WaitUntilImageIsActive(&n,&seed)
	
	if ( !ImageActive ) {
		log.Fatal("Seed Image is not active")
		os.Exit(1)
	}
	seed, _ = ntnxAPI.GetImagebyName(&n, seed.Name)

	//check if VM exists
	exist, _ := ntnxAPI.VMExist(&n, v.Config.Name)

	if ( exist ) {
		log.Warn("VM " + v.Config.Name + " already exists")
	} else {

		// Create VM
		taskUUID, _ = ntnxAPI.CreateVM_AHV(&n, &v)

		task, err := ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
		if err != nil {
			log.Fatal("Task does not exist")
		} else {
			log.Info("VM " + v.Config.Name + " created")
		}

		// Clone Cloud-Image disk
		v.UUID = ntnxAPI.GetVMIDbyTask(&n, &task)

		taskUUID, _ = ntnxAPI.CloneDiskforVM(&n, &v, &im)

		task, err = ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
		if err != nil {
			log.Fatal("Task does not exist")
		} else {
			log.Info("Disk ID" + v.UUID + " cloned")
		}

		// Clone Seed.iso to CDROM
		taskUUID, _ = ntnxAPI.CloneCDforVM(&n, &v, &seed)

		task, err = ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
		if err != nil {
			log.Fatal("Task does not exist")
		} else {
			log.Info("CD ISO ID" + v.UUID + " cloned")
		}

		//	Create Nic1
		nic1.UUID = ntnxAPI.GetNetworkIDbyName(&n, "VLAN0")

		taskUUID, _ = ntnxAPI.CreateVNicforVM(&n, &v, &nic1)

		task, err = ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
		if err != nil {
			log.Fatal("Task does not exist")
		} else {
			log.Info("Nic1 created")
		}

		//	Start Cloud-VM

		taskUUID, _ = ntnxAPI.StartVM(&n, &v)

		task, err = ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
		if err != nil {
			log.Fatal("Task does not exist")
		} else {
			log.Info("VM started")
		}

	}
}
