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

## MAIN block-------------------------------------------------------------

LOGDIR=""
AppVersion="1.0 stable"
HELP=0
VERSION=0
HOST=0
PASSWORD=0


printHelp()
{
cat << EOF
  USAGE:
    daily_health_report.sh [options] [value]
    create a daily health report and send it via sendEmail

  Options:
    --host        specifies Clustre or CVM IP
    --password    specifies the PRISM admin password
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

if [ $PASSWORD = 0 ]; then
 echo "--password is mandatory"
 exit
fi


## start logic

curl --user admin:$PASSWORD --insecure -H "Content-Type: application/json" -H "Accept: application/json" https://$HOST:9440/console/downloads/ncli.zip > ncli.zip
if [ $? -ne 0 ]; then
 echo "The ncli.zip download from $host failed"
 exit
fi

sudo rm -Rr /usr/local/ncli
sudo unzip ncli.zip -d /usr/local/ncli > /dev/null
