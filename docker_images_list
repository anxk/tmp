#!/bin/bash

# This script will archive images.

sed -i "/^#/d" ${1}
sed -i "/^$/d" ${1}

while read item; do
    docker pull ${item}
    archiveName=$(echo ${item} | tr ':/' '_')
    # use gzip to reduce the size of archives
    docker save ${item} | gzip > ${archiveName}.tar.gz
    split -n 20 ${archiveName}.tar.gz TTT
done < ${1}
