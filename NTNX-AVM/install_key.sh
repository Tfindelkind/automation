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

ssh-keygen -b 2048 -t rsa -f /home/nutanix/.ssh/id_rsa -q -N ""

ncli -s 192.168.178.130 -u admin -p nutanix/4u cluster status | grep Name | cut -d':' -f2 > cvm_list

while IFS='' read -r line || [[ -n "$line" ]]; do
     scp /home/nutanix/.ssh/id_rsa.pub nutanix@$line:/home/nutanix/
     echo $line
     ssh nutanix@$line 'cat id_rsa.pub >> .ssh/authorized_keys2'
done < cvm_list
