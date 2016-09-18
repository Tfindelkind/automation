name=$(date '+%y-%m-%d')

ssh nutanix@192.168.178.130 'nodetool -h 0 ring > daily_health-report-$name.txt' < /dev/null
ssh nutanix@192.168.178.130 'genesis status >> daily_health-report-$name.txt' < /dev/null
ssh nutanix@192.168.178.130 'cluster status >> daily_health-report-$name.txt' < /dev/null
ssh nutanix@192.168.178.130 'allssh df -h >> daily_health-report-$name.txt' < /dev/null
ssh root@192.168.1.1 'ncli alerts ls >> daily_health-report-$name.txt' < /dev/null
ssh root@192.168.1.1 'allssh "ls -lahrt ~/data/logs | grep -i fatal" >> daily_health-report-$name.txt' < /dev/null
