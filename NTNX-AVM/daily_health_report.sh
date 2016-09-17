
ssh nutanix@192.168.178.130 'nodetool -h 0 ring'
ssh root@192.168.1.1 'genesis status'
ssh root@192.168.1.1 'cluster status'
ssh root@192.168.1.1 'uptime' allssh df -h
ssh root@192.168.1.1 'uptime' ncli alerts ls
ssh root@192.168.1.1 'uptime' allssh "ls -lahrt ~/data/logs | grep -i fatal"
