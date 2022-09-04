# Streamer
A better streaming solution

The code here is split into three different sections: forwarder, recorder and server
These are named for the different aspects of this program and how it can be distributed between multiple servers
## Prerequisites
At least one server that can store video files and one server that can accept RTMP streams, i.e. nginx rtmp module (theoretically one server could do all of this although for performance, it is recommended that you have different servers for everything)

A user that can ssh to the forwarder and the recorder and execute the start and stop for each, and the user needs to be able to save files in the recording directory.
The forwarder and recorder will need a home directory that the files will be stored in
## How to build
Coming soon - `wget https://raw.githubusercontent.com/ystv/streamer/master/build/build.sh`

You need to build a total of 5 files in the correct locations, these include the `main.go` file in the server folder, the forwarder start and stop in the forwarder folder and the recorder start and stop in the recorder folder
### Building
#### Forwarder
`go build -o forwarder_start forwarder_start.go && go build -o forwarder_stop forwarder_stop.go`

This must be done in the directory of the files
#### Recorder
`go build -o recorder_start recorder_start.go && go build -o recorder_stop recorder_stop.go`

This must be done in the directory of the files
#### Server
`go build -o streamer main.go`

This must be done in the directory of the file

In addition, a `streams.db` file will need to be created or use the given file. it's a SQLite schema and can be created with this statement:

```
create table stored
(
stream    TEXT not null
unique,
input     TEXT not null,
recording TEXT,
website   TEXT,
streams   TEXT not null
);

create table streams
(
stream    TEXT                  not null
unique,
recording integer default true  not null,
website   integer default false not null,
streams   integer default 1     not null
);
```