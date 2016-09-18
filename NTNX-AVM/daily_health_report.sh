#!/bin/bash
#
# Copyright (c) 2016 Thomas Findelkind
#
# This program is free software: you can redistribute it and/or modify it under
# the terms of the GNU General Public License as published by the Free Software
# Foundation, either version 3 of the License, or (at your option) any later
# version.
#
# This program is distributed in the hope that it will be useful, but WITHOUT
# ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
# FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
# details.
#
# You should have received a copy of the GNU General Public License along with
# this program.  If not, see <http://www.gnu.org/licenses/>.
#
# MORE ABOUT THIS SCRIPT AVAILABLE IN THE README AND AT:
#
# http://tfindelkind.com
#
# ---------------------------------------------------------------------------- #
LOGDIR=""
AppVersion="1.0 stable"
HELP=0
VERSION=0
HOST=0
RECIPIENT=0
SMTPUser=0
SMTPPassword=0
SERVER=0
PORT=0


printHelp()
{
cat << EOF
  USAGE:
    daily_health_report.sh [options] [value]
    create a daily health report and send it via sendEmail

  Options:
    --host        specifies the Nutanix cluster IP or CVM IP
    --password    (Optional) spefifies the nutanix PASSWORD
    --help        list this help
    --version     shows the version of daily_health_report.sh
EOF
}

## MAIN block-------------------------------------------------------------

## parse parameter
for i in "$@"
do
case $i in
    --host=*)
    HOST="${i#*=}"
    shift # past argument=value
    ;;
    --password=*)
    PASSWORD="${i#*=}"
    shift # past argument=value
    ;;
    --help)
    HELP=1
    shift # past argument=value
    ;;
    --version)
    VERSION=1
    shift # past argument=value
    ;;
esac
done

## evaluate parameter
if [ $VERSION = 1 ]; then
 echo $AppVersion
 exit
fi

if [ $HELP = 1 ]; then
 printHelp
 exit
fi

if [ $HOST = 0 ]; then
 echo "--host is mandatory"
 exit
fi

## start logic

name=$(date '+%y-%m-%d')

ssh nutanix@192.168.178.132 "export PS1='fake>' ; source /etc/profile ; nodetool -h 0 ring > daily_health_report-$name.txt" < /dev/null
ssh nutanix@192.168.178.132 "export PS1='fake>' ; source /etc/profile ; genesis status >> daily_health_report-$name.txt" < /dev/null
ssh nutanix@192.168.178.132 "export PS1='fake>' ; source /etc/profile ; cluster status >> daily_health_report-$name.txt" < /dev/null
ssh nutanix@192.168.178.132 "export PS1='fake>' ; source /etc/profile ; df -h >> daily_health_report-$name.txt" < /dev/null
ssh nutanix@192.168.178.132 "export PS1='fake>' ; source /etc/profile ; ncli alerts ls >> daily_health_report-$name.txt" < /dev/null
ssh nutanix@192.168.178.132 "export PS1='fake>' ; source /etc/profile ; __allssh 'ls -lahrt ~/data/logs | grep -i fatal' >> daily_health_report-$name.txt" < /dev/null
