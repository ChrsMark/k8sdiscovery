#!/bin/bash
rm -rf profiles
mkdir -p profiles
sleepTime=$((60*5))
for i in $(seq 1 1 100)
do
   echo "Getting heap for $i time"
   go tool pprof -png http://localhost:6060/debug/pprof/heap > profiles/heap${i}.png
   sleep $sleepTime
done