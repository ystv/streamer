#!/bin/bash
if [ "$#" -ne 3 ]; then
	raise error "Illegal number of arguments!"
else
	i=0
	for (( ; ; ))
	do
		ffmpeg -i "$1" -c copy "$2_$i.mkv" >> "$3.txt" 2>&1
		((i=i+1))
		sleep 1
	done
fi