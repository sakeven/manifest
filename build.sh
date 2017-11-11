#!/bin/bash

work_dir=$(dirname $0)
cd ${work_dir}

if [ ! -d "bin" ]; then
    mkdir bin
fi

key="manifest"
echo building ${key} ...
go build -ldflags "${ldflags}" -o bin/${key} 
if [ $? -ne 0 ]; then
    echo build ${key} failed!
    exit 1
fi

echo success
