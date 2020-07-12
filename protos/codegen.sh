#!/bin/bash

files=`ls *.proto`

protoc \
 -I/usr/local/include \
 -I. \
 -I$GOPATH/src \
 --go_out=:. \
 --go_opt=paths=source_relative \
 $files
