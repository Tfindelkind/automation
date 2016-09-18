name=$(date '+%y-%m-%d')

ssh nutanix@192.168.178.130 '/usr/local/nutanix/apache-cassandra/bin/nodetool -h 0 ring > daily_health-report-$name.txt' < /dev/null
ssh nutanix@192.168.178.130 '/usr/local/nutanix/cluster/bin/genesis status >> daily_health-report-$name.txt' < /dev/null
ssh nutanix@192.168.178.130 '/usr/local/nutanix/cluster/bin/cluster status >> daily_health-report-$name.txt' < /dev/null
ssh nutanix@192.168.178.130 'export PS1="fake>" ; source /etc/profile ; df -h >> daily_health-report-$name.txt' < /dev/null
ssh nutanix@192.168.178.130 '~/prism/cli/ncli alerts ls >> daily_health-report-$name.txt' < /dev/null
ssh nutanix@192.168.178.130 'export PS1="fake>" ; source /etc/profile ; __allssh "ls -lahrt ~/data/logs | grep -i fatal" >> daily_health-report-$name.txt' < /dev/null
