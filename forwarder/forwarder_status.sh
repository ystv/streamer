#!/bin/bash
if [ "$#" -ne 2 ]; then
	raise error "Illegal number of arguments!"
fi
tail -n 25 "logs/$1_$2.txt"
