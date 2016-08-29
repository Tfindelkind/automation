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
	ntnxAPI "github.com/Tfindelkind/ntnx-golang-client-sdk"

	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const appVersion = "0.9 beta"

var (
	host         *string
	username     *string
	password     *string
	vmName       *string
	newVMName    *string
	vdiskMapping *string
	container    *string
	listMapping  *bool
	keepImages   *bool
	debug        *bool
	delete       *bool
	overwrite    *bool
	help         *bool
	version      *bool
)

func init() {
	host = flag.String("host", "", "a string")
	username = flag.String("username", "", "a string")
	password = flag.String("password", "", "a string")
	vmName = flag.String("vm-name", "", "a string")
	newVMName = flag.String("new-vm-name", "", "a string")
	vdiskMapping = flag.String("vdisk-mapping", "", "a string")
	container = flag.String("container", "", "a string")
	listMapping = flag.Bool("list-mapping", false, "a bool")
	debug = flag.Bool("debug", false, "a bool")
	keepImages = flag.Bool("keep-images", false, "a bool")
	delete = flag.Bool("delete", false, "a bool")
	overwrite = flag.Bool("overwrite", false, "a bool")
	help = flag.Bool("help", false, "a bool")
	version = flag.Bool("version", false, "a bool")
}

func printHelp() {

	fmt.Println("Usage: move_vm [OPTIONS]")
	fmt.Println("move_vm [ --help | --version ]")
	fmt.Println("")
	fmt.Println("FOR NUTANIX AHV ONLY- clones a VM from one container to another")
	fmt.Println("vNic MAC Addresses will change unless --delete is used")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("")
	fmt.Println("--host             Specify CVM host or Cluster IP")
	fmt.Println("--username         Specify username for connect to host")
	fmt.Println("--password         Specify password for user")
	fmt.Println("--vm-name          Specify Virtual Machine name which will be moved")
	fmt.Println("--new-vm-name      New Virtual Machine name if specified (clone)")
	fmt.Println("--vdisk-mapping    Speficy the container mapping for each vdisk - ORDER IS IMPORTANT")
	fmt.Println("--list-mapping     Shows the actual vdisk-mapping")
	fmt.Println("--container        Specify the container where vm will be moved to")
	fmt.Println("--debug            Enables debug mode")
	fmt.Println("--keep-images      If enabled clones to image service will not be deleted")
	fmt.Println("--delete           Deletes soruce VM - ARE YOU REALLY SURE?")
	fmt.Println("--overwrite		Overwrites target VM/Images (delete and creates new one)")
	fmt.Println("--help             List this help")
	fmt.Println("--version          Show the deploy_cloud_vm version")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("")
	fmt.Println("move_vm --host=NTNX-CVM --username=admin --password=nutanix/4u --vm-name=MyVM --image-container=ISO")
	fmt.Println("move_vm --host=NTNX-CVM --username=admin --password=nutanix/4u --vm-name=MyVM --vdisk-mapping=scsi.0/ISO,scsi.1/Prod2")
	fmt.Println("")
}

// parse --vdisk-mapping or --container and checks if all container exist
func parseVdiskMapping(n *ntnxAPI.NTNXConnection) ([]ntnxAPI.VMDisks, error) {

	defer func() {
		if err := recover(); err != nil {
			log.Fatal("--vdisk-mapping seems not to have right format")
			os.Exit(1)
		}
	}()

	var vdiskMappings []ntnxAPI.VMDisks

	var VMDisk ntnxAPI.VMDisks

	result := strings.Split(*vdiskMapping, ",")

	// add Mappings
	for i := range result {

		res := strings.Split(result[i], "/")

		resAddr := strings.Split(res[0], ".")

		VMDisk.Addr.DeviceBus = resAddr[0]
		VMDisk.Addr.DeviceIndex, _ = strconv.Atoi(resAddr[1])

		// check if right format is used
		if !(VMDisk.Addr.DeviceBus == "scsi" || VMDisk.Addr.DeviceBus == "pci" || VMDisk.Addr.DeviceBus == "ide") {
			log.Error("--vdisk-mapping seems not to have right format")
			os.Exit(1)
		}

		if !(VMDisk.Addr.DeviceIndex >= 0 && VMDisk.Addr.DeviceIndex <= 255) {
			log.Error("--vdisk-mapping seems not to have right format")
			os.Exit(1)
		}

		if res[1] != "EMPTY" {
			containerUUID, err := ntnxAPI.GetContainerUUIDbyName(n, res[1])
			if err != nil {
				os.Exit(1)
			}
			VMDisk.ContainerUUID = containerUUID
		}

		vdiskMappings = append(vdiskMappings, VMDisk)

	}

	return vdiskMappings, nil
}

func checkVdiskMapping(v ntnxAPI.VMJSONAHV, VdiskMapping []ntnxAPI.VMDisks) {

	defer func() {
		if err := recover(); err != nil {
			log.Fatal("--vdisk-mapping is not correct")
			os.Exit(1)
		}
	}()

	for i, elem := range v.Config.VMDisks {
		if elem.Addr.DeviceBus != VdiskMapping[i].Addr.DeviceBus || elem.Addr.DeviceIndex != VdiskMapping[i].Addr.DeviceIndex {
			log.Error("--vdisk-mapping some source vdisks are not mapped")
			os.Exit(1)
		}
	}

	if len(v.Config.VMDisks) != len(VdiskMapping) {
		log.Error("--vdisk-mapping some source vdisks are not mapped")
		os.Exit(1)
	}

}

func evaluateFlags() (ntnxAPI.NTNXConnection, ntnxAPI.VMJSONAHV, ntnxAPI.VMJSONAHV, []ntnxAPI.VMDisks) {

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

	//delete
	if *delete {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("THIS WILL DELETE source VM: " + *vmName)
		fmt.Print("If you want to continue type YES: ")
		text, _ := reader.ReadString('\n')
		fmt.Println(text)
		if strings.TrimRight(text, "\n") != "YES" {
			os.Exit(0)
		}
	}

	//host
	if *host == "" {
		log.Warn("mandatory option '--host=' is not set")
		os.Exit(0)
	}

	//username
	if *username == "" {
		log.Warn("mandatory option '--username=' is not set")
		os.Exit(0)
	}

	//password
	if *password == "" {
		log.Warn("mandatory option '--password=' is not set")
		os.Exit(0)
	}

	//vm-name
	if *vmName == "" {
		log.Warn("mandatory option '--vm-name=' is not set")
		os.Exit(0)
	}
	var vm ntnxAPI.VMJSONAHV
	vm.Config.Name = *vmName

	// new-vm-name
	if *newVMName == "" {
		newVMName = vmName
	}
	var vNew ntnxAPI.VMJSONAHV
	vNew.Config.Name = *newVMName

	var n ntnxAPI.NTNXConnection

	n.NutanixHost = *host
	n.Username = *username
	n.Password = *password

	ntnxAPI.EncodeCredentials(&n)
	ntnxAPI.CreateHTTPClient(&n)

	ntnxAPI.NutanixCheckCredentials(&n)

	// list mapping if specified
	if *listMapping {
		var listMappingStr string

		exist, _ := ntnxAPI.VMExist(&n, vm.Config.Name)
		if exist {
			vm, _ = ntnxAPI.GetVMbyName(&n, &vm)

			for i, elem := range vm.Config.VMDisks {
				if !elem.IsEmpty {
					containerName, _ := ntnxAPI.GetContainerNamebyUUID(&n, elem.ContainerUUID)
					listMappingStr = listMappingStr + elem.Addr.DeviceBus + "." + strconv.Itoa(elem.Addr.DeviceIndex) + "/" + containerName
				} else {
					listMappingStr = listMappingStr + elem.Addr.DeviceBus + "." + strconv.Itoa(elem.Addr.DeviceIndex) + "/EMPTY"
				}
				if i < len(vm.Config.VMDisks)-1 {
					listMappingStr = listMappingStr + ","
				}
			}
			fmt.Println(listMappingStr)
			os.Exit(0)
		}
	}

	// both options set container and vdisk-mapping
	if *container != "" && *vdiskMapping != "" {
		log.Warn("Option --container and --vdisk-mapping are set. Only one of them is allowed")
		os.Exit(0)
	}

	// none options set container and vdisk-mapping
	if *container == "" && *vdiskMapping == "" {
		log.Warn("None of --container or --vdisk-mapping is set. One is mandatory")
		os.Exit(0)
	}

	// If container is not found exit
	if *container != "" {
		_, err := ntnxAPI.GetContainerUUIDbyName(&n, *container)
		if err != nil {
			os.Exit(1)
		}

	}

	var VdiskMapping []ntnxAPI.VMDisks
	var err error
	// If container is not found exit
	if *vdiskMapping != "" {
		VdiskMapping, err = parseVdiskMapping(&n)
		if err != nil {
			os.Exit(1)
		}

	}

	return n, vm, vNew, VdiskMapping

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
	var vNew ntnxAPI.VMJSONAHV
	var d ntnxAPI.VDiskJSONREST
	var net ntnxAPI.NetworkREST
	var im ntnxAPI.ImageJSONAHV
	var taskUUID ntnxAPI.TaskUUID
	var VdiskMapping []ntnxAPI.VMDisks
	var existV bool
	var existVNew bool

	n, v, vNew, VdiskMapping = evaluateFlags()

	/*
	   Short description what will be done

	   1. Upload vDisk from Source VM to Image Service. This is needed while a direct copy is not possible and wait
	   2. Create VM and wait
	   3. Clone Images to Disks and wait
	   4. Add network
	   5. delete images
	*/

	/*To-DO:

	  2. show_progress


	*/

	//check if source VM exists
	existV, _ = ntnxAPI.VMExist(&n, v.Config.Name)

	if existV {

		//check if new VM exists
		existVNew, _ = ntnxAPI.VMExist(&n, vNew.Config.Name)

		if existVNew {
			log.Warn("VM " + vNew.Config.Name + " already exists")
			if !*overwrite {
				os.Exit(0)
			} else {
				vNew, _ = ntnxAPI.GetVMbyName(&n, &vNew)
			}

		}

		v, _ = ntnxAPI.GetVMbyName(&n, &v)

		// check if Source VM is running
		state := ntnxAPI.GetVMState(&n, &v)

		if state != "off" {
			log.Warn("VM " + vNew.Config.Name + " is not powered off")
			os.Exit(0)
		}

		//check if all disks have been specified
		if *vdiskMapping != "" {
			checkVdiskMapping(v, VdiskMapping)
		}

		var taskUUIDS []ntnxAPI.TaskUUID

		// clone vDisk from source VM to image service
		for i, elem := range v.Config.VMDisks {
			if !elem.IsEmpty {
				var containerName string
				d.VdiskUUID = elem.VMDiskUUID
				d.ContainerID = elem.ContainerUUID
				im.Name = v.Config.Name + "-" + elem.Addr.DeviceBus + "." + strconv.Itoa(elem.Addr.DeviceIndex)
				im.Annotation = "vm_move helper Image"

				// set containerName dependent on --container or --vdisk-mapping
				if *container != "" {
					containerName = *container
				} else {
					containerName, _ = ntnxAPI.GetContainerNamebyUUID(&n, VdiskMapping[i].ContainerUUID)
				}

				// make sure Images don't exist and overwrite if flag is enabled else WARN and continue
				// let all Upload take place in parallel and save taskUUID in a Array
				if ntnxAPI.ImageExistbyName(&n, &im) {
					if *overwrite {
						task, _ := ntnxAPI.DeleteImagebyName(&n, im.Name)
						ntnxAPI.WrappWaitUntilTaskFinished(&n, task.TaskUUID, "Previos existing Image "+im.Name+" deleted")

						taskUUID, _ = ntnxAPI.CreateImageFromURL(&n, &d, &im, containerName)
						taskUUIDS = append(taskUUIDS, taskUUID)
						log.Info("Start cloning " + im.Name + " to image service")
						// if overwrite is disabled
					} else {
						log.Info("Image " + im.Name + " already exists - will use existing one instead")
					}
				} else {

					// let all Upload take place in parallel and save taskUUID in a Array
					// if container stays the same clone local
					if VdiskMapping[i].ContainerUUID == elem.ContainerUUID {
						taskUUID, _ = ntnxAPI.CreateImageFromVdisk(&n, &d, &im)
						log.Info("Start cloning " + im.Name + " to image service")
					} else {
						taskUUID, _ = ntnxAPI.CreateImageFromURL(&n, &d, &im, containerName)
						log.Info("Start cloning " + im.Name + " to image service from URL")
					}
					taskUUIDS = append(taskUUIDS, taskUUID)

				}
			}
		}

		//Wait that all disks have been clone to new container- may take a while
		for _, task := range taskUUIDS {
			ntnxAPI.WrappWaitUntilTaskFinished(&n, task.TaskUUID, "Image from disk created")
		}

		//copy all VM settings
		vNew.Config.Description = v.Config.Description
		vNew.Config.MemoryMb = v.Config.MemoryMb
		vNew.Config.NumVcpus = v.Config.NumVcpus
		vNew.Config.NumCoresPerVcpu = v.Config.NumCoresPerVcpu

		// Delete target VM if overwrite mode
		if *overwrite && existVNew {

			taskUUID, _ = ntnxAPI.DeleteVM(&n, &vNew)

			ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID, "New VM successfull deleted")

		}

		// Create target VM
		taskUUID, _ = ntnxAPI.CreateVMAHV(&n, &vNew)

		task, err := ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)

		if err != nil {
			log.Fatal("Task does not exist")
		} else {
			log.Info("VM " + vNew.Config.Name + " created")
		}

		vNew.UUID = ntnxAPI.GetVMIDbyTask(&n, &task)

		// Create vdisk on new VM from images
		for _, elem := range v.Config.VMDisks {
			if elem.IsEmpty {
				taskUUID, _ = ntnxAPI.CreateCDforVMwithDetails(&n, &vNew, elem.Addr.DeviceBus, strconv.Itoa(elem.Addr.DeviceIndex))
				ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID, "CD successfully created")

			} else {
				im, _ = ntnxAPI.GetImagebyName(&n, v.Config.Name+"-"+elem.Addr.DeviceBus+"."+strconv.Itoa(elem.Addr.DeviceIndex))

				if elem.IsCdrom {
					taskUUID, _ = ntnxAPI.CloneCDforVMwithDetails(&n, &vNew, &im, elem.Addr.DeviceBus)
				} else {
					taskUUID, _ = ntnxAPI.CloneDiskforVMwithDetails(&n, &vNew, &im, elem.Addr.DeviceBus)
				}
				ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID, "Disk ID"+v.UUID+" cloned")
			}
		}

		// Create vNics
		for _, elem := range v.Config.VMNics {

			if *delete {
				// delete nic from source VM because only one NIC with same MAC may exist
				taskUUID, _ = ntnxAPI.DelteVNicforVM(&n, &v, elem.MacAddress)

				ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID, "Nic with MAC "+elem.MacAddress+" created")

				//	Create nic

				net.UUID = elem.NetworkUUID

				taskUUID, _ = ntnxAPI.CreateVNicforVMwithMAC(&n, &vNew, &net, elem.MacAddress)

				ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID, "Nic with MAC "+elem.MacAddress+" created")

			} else {
				net.UUID = elem.NetworkUUID

				taskUUID, _ = ntnxAPI.CreateVNicforVM(&n, &vNew, &net)

				ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID, "Nic with MAC "+elem.MacAddress+" created")

			}
		}

		// delete images
		if !*keepImages {
			for _, elem := range v.Config.VMDisks {
				if !elem.IsEmpty {

					d.VdiskUUID = elem.VMDiskUUID
					d.ContainerID = elem.ContainerUUID
					im.Name = v.Config.Name + "-" + elem.Addr.DeviceBus + "." + strconv.Itoa(elem.Addr.DeviceIndex)
					im.Annotation = "vm_move helper Image"

					task, _ := ntnxAPI.DeleteImagebyName(&n, im.Name)
					ntnxAPI.WrappWaitUntilTaskFinished(&n, task.TaskUUID, "Image "+im.Name+" deleted")
				}
			}
		}

		// delete VM if flag is set
		if *delete {

			taskUUID, _ = ntnxAPI.DeleteVM(&n, &v)

			ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID, "VM "+v.Config.Name+" successfull deleted")

		}

	} else {

		log.Warn("VM vm-name=" + v.Config.Name + " does not exist")

	}

}
