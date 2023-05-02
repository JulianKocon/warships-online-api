#!/bin/bash

i=0
while true;
do
    name=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 6 ; echo '')
    i=$((i+1))
    echo "Win streak: $i"
    go run main.go --nick "$name" --abandon &> /dev/null &
    sleep 0.5
    go run main.go --nick JulianKocon --target_nick "$name" &> /dev/null
    wait
done    