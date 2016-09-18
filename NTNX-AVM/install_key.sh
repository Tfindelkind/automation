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
PASSWORD=0
HOST=0



printHelp()
{
cat << EOF
  USAGE:
    install_key.sh [options] [value]
    create keys and deploys them to the Nutanix CVMs via SCP/SSH

  Options:
    --host        specifies the Nutanix cluster IP or CVM IP
    --password    spefifies the Prism admin PASSWORD
    --help        list this help
    --version     shows the version of install_key.sh
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

if [ $PASSWORD = 0 ]; then
 echo "--password is mandatory"
 exit
fi

## start logic

## generate new keypair without pass
ssh-keygen -b 2048 -t rsa -f /home/nutanix/.ssh/id_rsa -q -N ""

ncli -s $HOST -u admin -p "$PASSWORD" cluster status | grep Name | cut -d':' -f2 | tr -d ' ' > cvm_list

echo "You need to enter the ssh password for each CVM two times."
echo ""


## deploy keys to server
while IFS='' read -r line || [[ -n "$line" ]]; do
       echo "scp public key for $line"
       scp /home/nutanix/.ssh/id_rsa.pub nutanix@$line:/home/nutanix/
       echo "add public key for $line"
       ssh nutanix@$line 'cat id_rsa.pub >> .ssh/authorized_keys2; rm id_rsa.pub' < /dev/null
done < cvm_list
