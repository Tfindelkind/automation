# move_vm

mv_vm - golang binary which leverages the Nutanix REST API to move a VM from one container to another one (AKA SVMOTION)

!!! Nutanix AHV ONLY - SOURCE VM MUST BE STOPPED-> no online motion !!! 

This is related to the fact that Nutanix don't provide an online storage motion atm. 
The Nutanix KB 000002663 shows the manual tasks which are involved to copy a vm from one container to another

In the first release I am using the unofficial ntnxAPI SDK but in future release I will change to the official Nutanix Golang SDK.

# Installing

I recommend to install the Automation VM for an easy use. 

go get https://github.com/Tfindelkind/move_vm

# Usage

move_vm --host=NTNX-CVM --username=admin --password=nutanix/4u --vm-name=MyVM --image-container=prod

