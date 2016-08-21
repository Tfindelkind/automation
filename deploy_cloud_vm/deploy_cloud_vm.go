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
  host *string
  username *string
  password *string
  vm_name *string
  image_name *string
  seed_name *string
  image_file *string
  seed_file *string
  vlan *string
  image_container *string
  vm_container *string
  debug *bool
  help *bool
  version *bool
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
  image_container = flag.String("image-container", "ISO", "a string")
  vm_container = flag.String("vm-container", "prod", "a string")
  debug = flag.Bool("debug", false, "a bool")
  help = flag.Bool("help", false, "a bool")
  version = flag.Bool("version", false, "a bool")
}
	
func printHelp() {

 fmt.Println ("Usage: deploy_cloud_vm [OPTIONS]"); 
 fmt.Println ("deploy_cloud_vm [ --help | --version ]")
 fmt.Println ("")
 fmt.Println ("Upload and deploy a cloud image with a CD seed")
 fmt.Println ("Example seed.iso at https://github.com/Tfindelkind/DCI")
 fmt.Println ("")
 fmt.Println ("Options:")
 fmt.Println ("")
 fmt.Println ("--host")
 fmt.Println ("--username")
 fmt.Println ("--password")
 fmt.Println ("--vm-name")
 fmt.Println ("--image-name")
 fmt.Println ("--image-file")
 fmt.Println ("--seed-name")
 fmt.Println ("--seed-file")
 fmt.Println ("--vlan")
 fmt.Println ("--image_container")
 fmt.Println ("--vm_container")
 fmt.Println ("--debug")
 fmt.Println ("--help")
 fmt.Println ("--debug")
 fmt.Println ("")
 fmt.Println ("Example:")
 fmt.Println ("")
 fmt.Println ("deploy_cloud_vm --host=NTNX-CVM --username=admin --password=nutanix/4u --vm-name=NTNX-AVM --image-name=Centos7-1606 --image-container=ISO --vm-container=prod vlan=VLAN0") 
  
}	
	

func main() {	
	
	flag.Parse()	
	
	if ( *help ) {
	 printHelp()
	 os.Exit(0)
	} 
	
	if ( *version ) {
	 fmt.Println("Version: "+AppVersion)
	 os.Exit(0)
	} 
	
	if ( *debug ) {
	 log.SetLevel(log.DebugLevel)
	} else {
	 log.SetLevel(log.InfoLevel)
	}	
	customFormatter := new(log.TextFormatter)
    customFormatter.TimestampFormat = "2006-01-02 15:04:05"
    log.SetFormatter(customFormatter)
    customFormatter.FullTimestamp = true
	
	var n 			ntnxAPI.NTNXConnection
	var v 			ntnxAPI.VM_json_AHV
	var nic1		ntnxAPI.Network_REST
	var im			ntnxAPI.Image_json_AHV
	var seed		ntnxAPI.Image_json_AHV
	var taskUUID 	ntnxAPI.TaskUUID

	
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
	   1. commandline help
	   2. Inplement progress bar- (concurreny and get progress from task)
	    
	  
	  */ 
	
	// upload cloud image to image service
    if ( ntnxAPI.ImageExistbyName(&n,&im) ) {
	 log.Warn("Image "+im.Name+" already exists")
	 // get existing image ID 
	 im.UUID, _ = ntnxAPI.GetImageIDbyName(&n,im.Name)
	 } else {		
		taskUUID, _ = ntnxAPI.CreateImageObject(&n,&im)
	
		task, err := ntnxAPI.WaitUntilTaskFinished(&n,taskUUID.TaskUUID)
		if ( err != nil ) {
			log.Fatal("Task does not exist")	
		}
	
		im.UUID = ntnxAPI.GetImageIDbyTask(&n,&task)

		_ , statusCode := ntnxAPI.PutFileToImage(&n,ntnxAPI.NutanixAHVurl(&n),"images/"+im.UUID+"/upload",*image_file,*image_container)
		
		if (statusCode != 200) {
			log.Error("Image upload failed")
			os.Exit(1)	
		}
    }
   
   // upload seed.iso to image service
   if ( ntnxAPI.ImageExistbyName(&n,&seed) ) {
	 log.Warn("Image "+seed.Name+" already exists")	 
	 seed.UUID, _ = ntnxAPI.GetImageIDbyName(&n,seed.Name)
	} else {		
		taskUUID, _ = ntnxAPI.CreateImageObject(&n,&seed)
	
		task, err := ntnxAPI.WaitUntilTaskFinished(&n,taskUUID.TaskUUID)
		if ( err != nil ) {
			log.Fatal("Task does not exist")	
		}
	
		seed.UUID = ntnxAPI.GetImageIDbyTask(&n,&task)

		_ , statusCode := ntnxAPI.PutFileToImage(&n,ntnxAPI.NutanixAHVurl(&n),"images/"+seed.UUID+"/upload",*seed_file,*image_container)
		
		if (statusCode != 200) {
			log.Error("Image upload failed")
			os.Exit(1)	
		}
    }
    
    //check if VM exists
    exist, _ := ntnxAPI.VMExist(&n,v.Config.Name)
        
    if ( exist ) {
	 log.Warn("VM "+v.Config.Name+" already exists")	 
	} else {
    		
 	// Create VM
 	taskUUID, _ = ntnxAPI.CreateVM_AHV(&n, &v)
 	
 	task, err := ntnxAPI.WaitUntilTaskFinished(&n,taskUUID.TaskUUID)
		if ( err != nil ) {
			log.Fatal("Task does not exist")	
		} else {
		 log.Info("VM "+v.Config.Name+" created")
		} 
		
	
	// Clone Cloud-Image disk 
	v.UUID = ntnxAPI.GetVMIDbyTask(&n,&task)
	
	taskUUID, _ = ntnxAPI.CloneDiskforVM(&n,&v,&im)
	
	task, err = ntnxAPI.WaitUntilTaskFinished(&n,taskUUID.TaskUUID)
		if ( err != nil ) {
			log.Fatal("Task does not exist")	
		} else {
	     log.Info("Disk ID"+v.UUID+" cloned")		
		}	
		
	// Clone Seed.iso to CDROM
	taskUUID, _ = ntnxAPI.CloneCDforVM(&n,&v,&seed)
	
	task, err = ntnxAPI.WaitUntilTaskFinished(&n,taskUUID.TaskUUID)
		if ( err != nil ) {
			log.Fatal("Task does not exist")	
		} else {
	     log.Info("CD ISO ID"+v.UUID+" cloned")		
		}	
	
	//	Create Nic1
    nic1.UUID = ntnxAPI.GetNetworkIDbyName(&n,"VLAN0")
    
    taskUUID, _ = ntnxAPI.CreateVNicforVM(&n,&v,&nic1)
	
	task, err = ntnxAPI.WaitUntilTaskFinished(&n,taskUUID.TaskUUID)
		if ( err != nil ) {
			log.Fatal("Task does not exist")	
		} else {
		 log.Info("Nic1 created")
		}
		
	//	Start Cloud-VM
        
    taskUUID, _ = ntnxAPI.StartVM(&n,&v)
	
	task, err = ntnxAPI.WaitUntilTaskFinished(&n,taskUUID.TaskUUID)
		if ( err != nil ) {
			log.Fatal("Task does not exist")	
		} else {
		 log.Info("VM started")
		}	
		  	
	}
}
