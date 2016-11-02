#!/bin/bash

files=`ls *.proto`

protoc \
 -I/usr/local/include \
 -I. \
 -I$GOPATH/src \
 --go_out=:. \
 $files
