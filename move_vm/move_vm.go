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
	"flag"
	"os"
	"strconv"
	)
	
	
const AppVersion = "0.9 beta"

var (
	host            *string
	username        *string
	password        *string
	vm_name         *string
	new_vm_name     *string
	vdisk_mapping   *string
	container       *string
	debug           *bool
	delete          *bool	
	overwrite		*bool
	help            *bool
	version         *bool
)
	
func init() {
	host = flag.String("host", "192.168.178.130", "a string")
	username = flag.String("username", "admin", "a string")
	password = flag.String("password", "nutanix/4u", "a string")
	vm_name = flag.String("vm-name", "MyVM", "a string")
	new_vm_name = flag.String("new-vm-name", *vm_name+"_copy", "a string")
	vdisk_mapping = flag.String("vdisk-mapping", "", "a string")
	container = flag.String("container", "ISO", "a string")	
	debug = flag.Bool("debug", false, "a bool")
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
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("")
	fmt.Println("--host             Specify CVM host or Cluster IP")
	fmt.Println("--username         Specify username for connect to host")
	fmt.Println("--password         Specify password for user")
	fmt.Println("--vm-name          Specify Virtual Machine name which will be moved")
	fmt.Println("--new-vm-name      New Virtual Machine name if specified (clone)")
	fmt.Println("--vdisk-mapping    Speficy the container mapping for each file")
	fmt.Println("--container        Specify the container where vm will be moved to")
	fmt.Println("--debug            Enables debug mode")
	fmt.Println("--delete           Deletes soruce VM - ARE you really sure?")
	
	fmt.Println("--help             List this help")
	fmt.Println("--version          Show the deploy_cloud_vm version")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("")
	fmt.Println("move_vm --host=NTNX-CVM --username=admin --password=nutanix/4u --vm-name=MyVM --image-container=ISO")
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
	
	var n 		 ntnxAPI.NTNXConnection
	var v 		 ntnxAPI.VM_json_AHV
	var v_new	 ntnxAPI.VM_json_AHV		 
	var d		 ntnxAPI.VDisk_json_REST
	var net 	 ntnxAPI.Network_REST
	var im 		 ntnxAPI.Image_json_AHV
	var taskUUID ntnxAPI.TaskUUID
	
	
	n.NutanixHost = *host
	n.Username = *username
	n.Password = *password
	v.Config.Name = *vm_name
	v_new.Config.Name = *new_vm_name
	
	ntnxAPI.EncodeCredentials(&n)
	ntnxAPI.CreateHttpClient(&n)
	
	/*
	   Short description what will be done

	   1. Upload vDisk from Source VM to Image Service. This is needed while a direct copy is not possible and wait
	   2. Create VM and wait
	   3. Clone Images to Disks and wait
	   4. Add network
	   5. delete images
	*/

	/*To-DO:
	
	  1. Check if Images exist already!
	  2. show_progress
	  

	*/
	
	//check if source VM exists
	exist, _ := ntnxAPI.VMExist(&n, v.Config.Name)

	if ( exist ) {
		
		//check if new VM exists
		exist, _ := ntnxAPI.VMExist(&n, v_new.Config.Name)

		if ( exist ) {
			log.Warn("VM " + v_new.Config.Name + " already exists")
			os.Exit(0)
		} 
				
		v, _ = ntnxAPI.GetVMbyName(&n,&v) 
        
        var taskUUID_s []ntnxAPI.TaskUUID
        
		for _, elem := range v.Config.VMDisks {
			if !elem.IsCdrom { 				
				d.VdiskUUID = elem.VMDiskUUID
				d.ContainerID=elem.ContainerUUID 
				im.Name = v.Config.Name+"-"+elem.Addr.DeviceBus+"."+strconv.Itoa(elem.Addr.DeviceIndex)
				im.Annotation =  "vm_move helper Image"
				
				// make sure Images don't exist and overwrite if flag is enabled else WARN and continue
				// let all Upload take place in parallel and save taskUUID in a Array
				if ntnxAPI.ImageExistbyName(&n, &im) {
					if *overwrite {		
							task, _ := ntnxAPI.DeleteImagebyName(&n,im.Name)  
							_ , err := ntnxAPI.WaitUntilTaskFinished(&n, task.TaskUUID)
							if err != nil {
								log.Fatal("Task does not exist")
							} else {
							  log.Info("Previos existing Image "+im.Name+" deleted")
							}
							taskUUID, _ = ntnxAPI.CreateImageFromURL(&n,&d,&im,*container)				
							taskUUID_s = append(taskUUID_s,taskUUID)
							log.Info("Start cloning " + im.Name + " to image service")
					// if overwrite is disabled 
					} else {
						log.Info("Image " + im.Name + " already exists - will use existing one instead")							
					}
				} else {					
					
				// let all Upload take place in parallel and save taskUUID in a Array
				taskUUID, _ = ntnxAPI.CreateImageFromURL(&n,&d,&im,*container)				
				taskUUID_s = append(taskUUID_s,taskUUID)
				log.Info("Start cloning " + im.Name + " to image service")
				}
			}
		 }	
		  		
		  //Wait that all disks have been clone to new container- may take a while	  
		  for  _ , task := range taskUUID_s {	  
			_ , err := ntnxAPI.WaitUntilTaskFinished(&n, task.TaskUUID)
			if err != nil {
			 log.Fatal("Task does not exist")
			} else {
			 log.Info("Image from disk created")
			}
          }	
          
        //copy all VM settings
		v_new.Config.Description = v.Config.Description
		v_new.Config.MemoryMb = v.Config.MemoryMb
		v_new.Config.NumVcpus = v.Config.NumVcpus
		v_new.Config.NumCoresPerVcpu = v.Config.NumCoresPerVcpu

		// Create VM
		taskUUID, _ = ntnxAPI.CreateVM_AHV(&n, &v_new)

		task, err := ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
		if err != nil {
			log.Fatal("Task does not exist")
		} else {
			log.Info("VM " + v_new.Config.Name + " created")
		}

		for _, elem := range v.Config.VMDisks {
			if !elem.IsCdrom { 							 	
				im, _ = ntnxAPI.GetImagebyName(&n,v.Config.Name+"-"+elem.Addr.DeviceBus+"."+strconv.Itoa(elem.Addr.DeviceIndex))
				// Clone Cloud-Image disk
				v_new.UUID = ntnxAPI.GetVMIDbyTask(&n, &task)

				taskUUID, _ = ntnxAPI.CloneDiskforVM(&n, &v_new, &im)

				task, err = ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
				if err != nil {
					log.Fatal("Task does not exist")
				} else {
					log.Info("Disk ID" + v.UUID + " cloned")
				}
			}	
		}	
			
		for _, elem := range v.Config.VMNics {
			//	Create nics
			net.UUID = elem.NetworkUUID
			taskUUID, _ = ntnxAPI.CreateVNicforVMwithMAC(&n, &v_new,&net,elem.MacAddress)

			task, err = ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
			if err != nil {
				log.Fatal("Task does not exist")
			} else {
				log.Info("Nic with MAC "+elem.MacAddress+" created")
			}
		}	

	 }
              	 
           
          
 }
        
		 
	/*if (ntnxAPI.ImageExist(&n,&im)) {
		fmt.Println("Image exists") } 

		 	 
		 
	d.ContainerID = ntnxAPI.GetContainerIDbyName(&n,d.Name)
	
	im = ntnxAPI.Image { "boot2docker", "", "ISO_IMAGE", "",  ntnxAPI.GetImageIDbyName(&n,"boot2docker")}
	
		
	ntnxAPI.CloneCDforVM(&n,&v,&im)
	
	
	ntnxAPI.CreateVDiskforVM(&n,&v,&d)
	 
	nic1.UUID = ntnxAPI.GetNetworkIDbyName(&n,nic1.Name)
	ntnxAPI.CreateVNicforVM(&n, &v,&nic1)
	
	ntnxAPI.StartVM(&n,&v)
	
	for i:= 0; i < 100; i++ {
					
			fmt.Println(ntnxAPI.GetVMIP(&n,&v))
			time.Sleep(1000*time.Millisecond)*/
	

