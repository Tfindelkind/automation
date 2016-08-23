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
	"bufio"
	"strings"
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
	list_mapping	*bool
	keep_images		*bool
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
	new_vm_name = flag.String("new-vm-name", "", "a string")
	vdisk_mapping = flag.String("vdisk-mapping", "", "a string")
	container = flag.String("container", "", "a string")
	list_mapping = flag.Bool("list-mapping", false, "a bool")	
	debug = flag.Bool("debug", false, "a bool")
	keep_images = flag.Bool("keep-images", false, "a bool")
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
func parseVdiskMapping(n *ntnxAPI.NTNXConnection) ([]ntnxAPI.VMDisks,error) {
	
  defer func() {
	if err := recover(); err != nil {
	    log.Fatal("--vdisk-mapping seems not to have right format")
	    os.Exit(1)	   
	}
   }()

   var vdiskMapping []ntnxAPI.VMDisks
				
		
	var VMDisk ntnxAPI.VMDisks		
				
	result := strings.Split(*vdisk_mapping, ",")
				
	// add Mappings
	for i := range result {
			
	   res := strings.Split(result[i], "/")
		   
	   resAddr := strings.Split(res[0], ".")
			
		VMDisk.Addr.DeviceBus = resAddr[0]
		VMDisk.Addr.DeviceIndex, _ = strconv.Atoi(resAddr[1])
				
		// check if right format is used
		if ( !(VMDisk.Addr.DeviceBus == "scsi" || VMDisk.Addr.DeviceBus == "pci" || VMDisk.Addr.DeviceBus == "ide") ){
			log.Error("--vdisk-mapping seems not to have right format")
			os.Exit(1)
		}
								
		if ( !(VMDisk.Addr.DeviceIndex >= 0 && VMDisk.Addr.DeviceIndex <= 255) ) {
			log.Error("--vdisk-mapping seems not to have right format")
			os.Exit(1)
		}
				
		if ( res[1] != "EMPTY" ) {
			containerUUID, err := ntnxAPI.GetContainerUUIDbyName(n,res[1])
			if ( err != nil) {
				os.Exit(1)
			}		
			VMDisk.ContainerUUID = containerUUID	    			   
		}				    	
		   
		vdiskMapping = append(vdiskMapping,VMDisk)
			 
	}
		
  return vdiskMapping, nil
}

func checkVdiskMapping(v ntnxAPI.VM_json_AHV, VdiskMapping []ntnxAPI.VMDisks) {
	
	defer func() {
	if err := recover(); err != nil {
	    log.Fatal("--vdisk-mapping is not correct")
	    os.Exit(1)	   
	}
   }()	
	
	for i, elem := range v.Config.VMDisks {
		if ( elem.Addr.DeviceBus != VdiskMapping[i].Addr.DeviceBus || elem.Addr.DeviceIndex != VdiskMapping[i].Addr.DeviceIndex ) {				
				log.Error("--vdisk-mapping some source vdisks are not mapped")	
				os.Exit(1)
			}
	}				
		
	if ( len(v.Config.VMDisks) != len(VdiskMapping) ) {
		log.Error("--vdisk-mapping some source vdisks are not mapped")	
		os.Exit(1)
	}
	
}

func evaluateFlags() (ntnxAPI.NTNXConnection,ntnxAPI.VM_json_AHV,ntnxAPI.VM_json_AHV,[]ntnxAPI.VMDisks) {
	
	//help
	if *help {
		printHelp()
		os.Exit(0)
	}
	
    //version
	if *version {
		fmt.Println("Version: " + AppVersion)
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
		fmt.Println("THIS WILL DELETE source VM: "+*vm_name)
		fmt.Print("If you want to continue type YES: ")
		text, _ := reader.ReadString('\n')
		fmt.Println(text)
		if ( strings.TrimRight(text, "\n") != "YES" ) {
			os.Exit(0)
		}
	}	
	
	
	//host
	if ( *host == "" ) {
		log.Warn("mandatory option 'host' is not set")	
		os.Exit(0)
	}
	
	//username
	if ( *username == "" ) {
		log.Warn("mandatory option 'username' is not set")	
		os.Exit(0)
	}
	
	//password
	if ( *password == "" ) {
		log.Warn("mandatory option 'password' is not set")	
		os.Exit(0)
	}
	
	//vm-name	
	if ( *vm_name == "" ) {
		log.Warn("mandatory option 'vm-name' is not set")	
		os.Exit(0)
	}
	var v 		 		ntnxAPI.VM_json_AHV	
	v.Config.Name = *vm_name
	
	// new-vm-name
	if ( *new_vm_name == "" ) {
		new_vm_name = vm_name
	}			
	var v_new	 		ntnxAPI.VM_json_AHV	
	v_new.Config.Name = *new_vm_name
	
	var n 		 		ntnxAPI.NTNXConnection
	
	n.NutanixHost = *host
	n.Username = *username
	n.Password = *password
	
	ntnxAPI.EncodeCredentials(&n)
	ntnxAPI.CreateHttpClient(&n)
	
	// list mapping if specified
	if (*list_mapping) {
		var list_mapping_str string
		
		exist , _ := ntnxAPI.VMExist(&n, v.Config.Name)
		if ( exist ) {
			v, _ := ntnxAPI.GetVMbyName(&n,&v)
		
			for i, elem := range v.Config.VMDisks {
					if (!elem.IsEmpty) {
						containerName, _ :=  ntnxAPI.GetContainerNamebyUUID(&n,elem.ContainerUUID)
						list_mapping_str = list_mapping_str + elem.Addr.DeviceBus+"."+strconv.Itoa(elem.Addr.DeviceIndex)+"/"+containerName
					} else {
					    list_mapping_str = list_mapping_str + elem.Addr.DeviceBus+"."+strconv.Itoa(elem.Addr.DeviceIndex)+"/EMPTY"
					} 										
					if ( i < len(v.Config.VMDisks)-1 ) {
						list_mapping_str = list_mapping_str + ","
					}
			}
			fmt.Println(list_mapping_str)
			os.Exit(0)
		} 
	}	

	
	// both options set container and vdisk-mapping
	if ( *container != "" && *vdisk_mapping != "") {
		log.Warn("Option --container and --vdisk-mapping are set. Only one of them is allowed") 		
		os.Exit(0)
	}	
	
	// none options set container and vdisk-mapping
	if ( *container == "" && *vdisk_mapping == "") {
		log.Warn("None of --container or --vdisk-mapping is set. One is mandatory") 
		os.Exit(0)
	}
		
	
	// If container is not found exit
	if ( *container != "") {
		_ , err := ntnxAPI.GetContainerUUIDbyName(&n,*container)
		if ( err != nil) {
			os.Exit(1)
		}
		
	}
	
	var VdiskMapping 	[]ntnxAPI.VMDisks
	var err error
	// If container is not found exit
	if ( *vdisk_mapping != "") {			
		VdiskMapping, err = parseVdiskMapping(&n)	
		if ( err != nil) {
			os.Exit(1)
		}
		
	}
	
	return n,v,v_new,VdiskMapping
					
}


func main() {	
	
	flag.Usage = printHelp
	flag.Parse()	
	
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	
	var n 		 		ntnxAPI.NTNXConnection
	var v 		 		ntnxAPI.VM_json_AHV
	var v_new	 		ntnxAPI.VM_json_AHV		 
	var d				ntnxAPI.VDisk_json_REST
	var net 	 		ntnxAPI.Network_REST
	var im 		 		ntnxAPI.Image_json_AHV
	var taskUUID 		ntnxAPI.TaskUUID
	var VdiskMapping 	[]ntnxAPI.VMDisks
	var exist_v			bool
	var exist_v_new		bool
	
	n, v, v_new, VdiskMapping = evaluateFlags()			 			
	
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
	exist_v, _ = ntnxAPI.VMExist(&n, v.Config.Name)


	if ( exist_v ) {
		
		//check if new VM exists
		exist_v_new, _ = ntnxAPI.VMExist(&n, v_new.Config.Name)		

		if ( exist_v_new ) {
			log.Warn("VM " + v_new.Config.Name + " already exists")
			if ( !*overwrite )  { 
				os.Exit(0)
			} else {
			  v_new, _ = ntnxAPI.GetVMbyName(&n,&v_new) 				
			}	
				
		} 
				
		v, _ = ntnxAPI.GetVMbyName(&n,&v) 								
		
		// check if Source VM is running
		state := ntnxAPI.GetVMState(&n,&v) 		
		
		if ( state != "off" ) {
			log.Warn("VM " + v_new.Config.Name + " is not powered off")
			os.Exit(0)
		}				
				
		//check if all disks have been specified 
		if ( *vdisk_mapping != "" ) {
			checkVdiskMapping(v,VdiskMapping)				
		 }	
        
        
        var taskUUID_s []ntnxAPI.TaskUUID
        
		// clone vDisk from source VM to image service
		for i, elem := range v.Config.VMDisks {
		   if ( !elem.IsEmpty ) {
			var containerName string 				
			d.VdiskUUID = elem.VMDiskUUID
			d.ContainerID=elem.ContainerUUID 
			im.Name = v.Config.Name+"-"+elem.Addr.DeviceBus+"."+strconv.Itoa(elem.Addr.DeviceIndex)
			im.Annotation =  "vm_move helper Image"					
				
			// set containerName dependent on --container or --vdisk-mapping
			if ( *container != "" ) {
				containerName =  *container
			} else {								
				containerName, _ =  ntnxAPI.GetContainerNamebyUUID(&n,VdiskMapping[i].ContainerUUID)					
			} 		
			
			// make sure Images don't exist and overwrite if flag is enabled else WARN and continue
			// let all Upload take place in parallel and save taskUUID in a Array
			if ntnxAPI.ImageExistbyName(&n, &im) {
				if *overwrite {		
						task, _ := ntnxAPI.DeleteImagebyName(&n,im.Name)  
						ntnxAPI.WrappWaitUntilTaskFinished(&n, task.TaskUUID,"Previos existing Image "+im.Name+" deleted") 
												
						taskUUID, _ = ntnxAPI.CreateImageFromURL(&n,&d,&im,containerName)				
						taskUUID_s = append(taskUUID_s,taskUUID)
						log.Info("Start cloning " + im.Name + " to image service")
				// if overwrite is disabled 
				} else {
					log.Info("Image " + im.Name + " already exists - will use existing one instead")							
				}
			} else {					
			
				
			// let all Upload take place in parallel and save taskUUID in a Array
			// if container stays the same clone local
			if ( VdiskMapping[i].ContainerUUID == elem.ContainerUUID ) 	{
				taskUUID, _ = ntnxAPI.CreateImageFromVdisk(&n,&d,&im)
				log.Info("Start cloning " + im.Name + " to image service")
			} else {
				taskUUID, _ = ntnxAPI.CreateImageFromURL(&n,&d,&im,containerName)	
				log.Info("Start cloning " + im.Name + " to image service from URL")
			}
			taskUUID_s = append(taskUUID_s,taskUUID)
			
			}
		   }	
		 }	
		  		
		  //Wait that all disks have been clone to new container- may take a while	  
		  for  _ , task := range taskUUID_s {	  
			ntnxAPI.WrappWaitUntilTaskFinished(&n, task.TaskUUID,"Image from disk created") 			
          }	
          
        //copy all VM settings
		v_new.Config.Description = v.Config.Description
		v_new.Config.MemoryMb = v.Config.MemoryMb
		v_new.Config.NumVcpus = v.Config.NumVcpus
		v_new.Config.NumCoresPerVcpu = v.Config.NumCoresPerVcpu
		
		// Delete target VM if overwrite mode
		if ( *overwrite && exist_v_new) {
			
			taskUUID, _ = ntnxAPI.DeleteVM(&n, &v_new)
			
			ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID,"New VM successfull deleted") 
			
			
		}

		// Create target VM
		taskUUID, _ = ntnxAPI.CreateVM_AHV(&n, &v_new)


		task, err := ntnxAPI.WaitUntilTaskFinished(&n, taskUUID.TaskUUID)
		
		
		if err != nil {
			log.Fatal("Task does not exist")
		} else {
			log.Info("VM " + v_new.Config.Name + " created")
		}

		v_new.UUID = ntnxAPI.GetVMIDbyTask(&n, &task)
			
		// Create vdisk on new VM from images
		for _, elem := range v.Config.VMDisks {
			if ( elem.IsEmpty) {
				taskUUID, _ = ntnxAPI.CreateCDforVMwithDetails(&n, &v_new,elem.Addr.DeviceBus,strconv.Itoa(elem.Addr.DeviceIndex))
				ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID,"CD successfully created") 
			
			} else {	
					im, _ = ntnxAPI.GetImagebyName(&n,v.Config.Name+"-"+elem.Addr.DeviceBus+"."+strconv.Itoa(elem.Addr.DeviceIndex))					
					
					if ( elem.IsCdrom ) {
						taskUUID, _ = ntnxAPI.CloneCDforVMwithDetails(&n, &v_new, &im,elem.Addr.DeviceBus)
					} else {	
						taskUUID, _ = ntnxAPI.CloneDiskforVMwithDetails(&n, &v_new, &im,elem.Addr.DeviceBus)
					}
					ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID,"Disk ID" + v.UUID + " cloned") 					
			}
		}	
		
		
		// Create vNics
		for _, elem := range v.Config.VMNics {
			
			
			if ( *delete ) {
				// delete nic from source VM because only one NIC with same MAC may exist	
				taskUUID, _ = ntnxAPI.DelteVNicforVM(&n, &v,elem.MacAddress)

				ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID,"Nic with MAC "+elem.MacAddress+" created")
								

				//	Create nic
						
				net.UUID = elem.NetworkUUID
			
				taskUUID, _ = ntnxAPI.CreateVNicforVMwithMAC(&n, &v_new,&net,elem.MacAddress)

				ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID,"Nic with MAC "+elem.MacAddress+" created")

			} else {
				net.UUID = elem.NetworkUUID
			
				taskUUID, _ = ntnxAPI.CreateVNicforVM(&n, &v_new,&net)
				
				ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID,"Nic with MAC "+elem.MacAddress+" created")
				
			}	
		}
		
		// delete images 
		if ( !*keep_images ) {
			for _, elem := range v.Config.VMDisks {
				if ( !elem.IsEmpty ) {
							
					d.VdiskUUID = elem.VMDiskUUID
					d.ContainerID=elem.ContainerUUID 
					im.Name = v.Config.Name+"-"+elem.Addr.DeviceBus+"."+strconv.Itoa(elem.Addr.DeviceIndex)
					im.Annotation =  "vm_move helper Image"					
										
					task, _ := ntnxAPI.DeleteImagebyName(&n,im.Name)  
					ntnxAPI.WrappWaitUntilTaskFinished(&n, task.TaskUUID,"Image "+im.Name+" deleted") 						
				}	
			}	
		}	
		
		
		// delete VM if flag is set
		if ( *delete ) {
			
			taskUUID, _ = ntnxAPI.DeleteVM(&n, &v)
			
			ntnxAPI.WrappWaitUntilTaskFinished(&n, taskUUID.TaskUUID,"VM "+v.Config.Name+" successfull deleted") 
			
			
		}				

	 } else {
		 
		 log.Warn("VM vm-name=" + v.Config.Name + " does not exist")
		 
	 }
              	 
           
          
 }
       
		 
	
	

