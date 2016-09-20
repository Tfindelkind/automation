env GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin/deploy_cloud_vm_mac github.com/Tfindelkind/automation/deploy_cloud_vm
env GOOS=linux GOARCH=amd64 go build -o ./bin/linux/deploy_cloud_vm_linux github.com/Tfindelkind/automation/deploy_cloud_vm
env GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin/export_vm_mac github.com/Tfindelkind/automation/export_vm
env GOOS=linux GOARCH=amd64 go build -o ./bin/linux/export_vm_linux github.com/Tfindelkind/automation/export_vm  
