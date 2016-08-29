# move_vm

move_vm - golang binary which leverages the Nutanix REST API to move a VM from one container to another one (AKA SVMOTION)

!!! Nutanix AHV ONLY - SOURCE VM MUST BE STOPPED-> no online motion !!! 

This is related to the fact that Nutanix don't provide an online storage motion atm. 
The Nutanix KB 000002663 shows the manual tasks which are involved to copy a vm from one container to another

In the first release I am using the unofficial ntnxAPI SDK but in future release I will change to the official Nutanix Golang SDK.

See http://tfindelkind.com for all details and step-by-step guides

# Dependencies
# Make sure to NFS whitelist the host/VM where you run move_vm

I recommend to install the Nutanix automation VM (AVM) for an easy use. 
Found at https://github.com/Tfindelkind/DCI

or

Install golang from https://golang.org/ if you don't use AVM or a binary.
Don't forget to set "GOPATH"

# Installing

Download the binary from https://github.com/Tfindelkind/automation/releases

or

go get https://github.com/Tfindelkind/automation/move_vm

The binary will be in "GOPATH\bin"

# Usage

move_vm --host=NTNX-CVM --username=admin --password=nutanix/4u --vm-name=MyVM --image-container=prod

