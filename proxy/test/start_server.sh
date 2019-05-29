#!/bin/bash

cd ../bin
./proxy_server -c ../test/dblist_define.json -l 0.0.0.0
cd ../test
