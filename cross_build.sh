env GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin/deploy_cloud_vm github.com/Tfindelkind/automation/deploy_cloud_vm 
env GOOS=linux GOARCH=amd64 go build -o ./bin/linux/deploy_cloud_vm github.com/Tfindelkind/automation/deploy_cloud_vm 
