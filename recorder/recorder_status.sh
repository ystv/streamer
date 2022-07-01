#!/bin/bash
if [ "$#" -ne 1 ]; then
	raise error "Illegal number of arguments!"
fi
tail -n 26 "logs/$1.txt"
