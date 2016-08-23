# move_vm

move_vm is an easy script 
In the first release I am using the unofficial ntnxAPI SDK but in future release I will change to the official Nutanix Golang SDK.

mv_vm - golang scripts which leverages the Nutanix REST API to move a VM from one container to another one (AKA SVMOTION)

It makes use of recipes which are defined before used. The recipes are simple text files right now because they are easy to use and enable me to reduce the over in the beginning. This may change in the future.

Installing

I recommend to install the Automation VM for an easy use. 

go get https://github.com/Tfindelkind/ntnx-golang-client-sdk 
go get https://github.com/Tfindelkind/move_vm

Usage.

move_vm VMname Destination_Container

