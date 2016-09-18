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
PASSWORD=0
RECIPIENT=0
PROVIDER=0
EMAILUSER=0
EMAILPASS=0
SERVER=0
PORT=0


printHelp()
{
cat << EOF
  USAGE:
    ncc_health_report.sh [options] [value]
    runs a 'ncc health_checks run_all' filters ERROR/FATAL and send email

  Options:
    --host        specifies ONE Nutanix CVM IP
    --password    specifies the Nutanix ssh password
    --recipient   specifies the email recipient
    --provider    specifies the email provider
    --emailuser   specifies the email user
    --emailpass   spefifies the email password
    --server      only needed when provider=other is used
    --port        only needed when provider=other is used
    --help        list this help
    --version     shows the version of ncc_health_report.sh
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
    --recipient=*)
    RECIPIENT="${i#*=}"
    shift # past argument=value
    ;;
    --provider=*)
    PROVIDER="${i#*=}"
    shift # past argument=value
    ;;
    --emailuser=*)
    EMAILUSER="${i#*=}"
    shift # past argument=value
    ;;
    --emailpass=*)
    EMAILPASS="${i#*=}"
    shift # past argument=value
    ;;
    --server=*)
    SERVER="${i#*=}"
    shift # past argument=value
    ;;
    --port=*)
    PORT="${i#*=}"
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

if [ $RECIPIENT = 0 ]; then
 echo "--recipient is mandatory"
 exit
fi

if [ $PROVIDER = 0 ]; then
 echo "--provider is mandatory"
 exit
fi

if [ $EMAILUSER = 0 ]; then
 echo "--emailuser is mandatory"
 exit
fi

if [ $EMAILPASS = 0 ]; then
 echo "--emailpass is mandatory"
 exit
fi

sendEmail --listprovider > provider.tmp

VALIDPROVIDER=0

while IFS='' read -r prov || [[ -n "$prov" ]]; do
  if [ "$prov" == "$PROVIDER" ]; then
    VALIDPROVIDER=1
  fi
done < provider.tmp
rm provider.tmp

if [ $VALIDPROVIDER = 0 ]; then
 echo "Provider unknown"
 sendEmail --listprovider
 exit
fi


## start logic

ssh nutanix@$HOST "export PS1='fake>' ; source /etc/profile ; ncc health_checks run_all" < /dev/null

scp nutanix@$HOST:/home/nutanix/data/logs/ncc-output-latest.log /home/nutanix

message=$(cat /home/nutanix/ncc-output-latest.log | grep 'ERR\|FAIL')+$(tail /home/nutanix/ncc-output-latest.log)

if [ "$PROVIDER" == "other" ]; then
  echo $PROVIDER
  sendEmail --recipient=$RECIPIENT --subject="daily_health_report-$name from NTNX-AVM" --message=$message --provider=$PROVIDER --server=$SERVER --port=$PORT --user=$EMAILUSER --password=$EMAILPASS --file=/home/nutanix/ncc-output-latest.log
else
  sendEmail --recipient=$RECIPIENT --subject="daily_health_report-$name from NTNX-AVM" --message=$message --provider=$PROVIDER --user=$EMAILUSER --password=$EMAILPASS --file=/home/nutanix/ncc-output-latest.log
fi
