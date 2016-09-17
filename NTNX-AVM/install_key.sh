ssh-keygen -b 2048 -t rsa -f /home/nutanix/.ssh/id_rsa -q -N ""

ncli -s 192.168.178.130 -u admin -p nutanix/4u ms ls | grep Name | cut -d':' -f2 > cvm_list

while IFS='' read -r line || [[ -n "$line" ]]; do
    if [ $line != "root" ]
    scp id_rsa.pub nutanix@$line:/home/nutanix/
    ssh nutanix@$line cat id_rsa.pub >> .ssh/authorized_keys2
done < cvm_list
