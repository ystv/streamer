#!/bin/bash
if [ "$#" -ne 4 ]; then
	raise error "Illegal number of arguments!"
else
	for (( ; ; ))
	do
		ffmpeg -i "$1" -c copy -f flv "$2" >> "$3_$4.txt" 2>&1
		sleep 1
	done
fi