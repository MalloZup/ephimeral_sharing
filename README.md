# ephimeral_sharing

# Desc/goals:

this repo was create for vefiying a race condition behaviour happening with volume.download and volume.upload call with libvirt.

The code is in a good shape however i did not focus on refactoring/design of it because it was a POC.
The main goal of this binary is to:

1) create libvirt volumes locally ,
2) write some content to this vols
3) then read from this volume the content and write it to tempFile.

# prerequisites:

0) you need  go version go1.11.2 linux/amd64
1) you have a default pool in libvirt
2) you have some disk space in `/var/lib/libvirt/images'

#### Howto run it

1) create a dir in `tmp/performance`
 this directory will be were the tmp files will be created.

2) run with `./ephimeral_sharing -numb 2`


This will run with 2 goroutines only.

Without any arg it will use 500 and create volume and files


3) REMOTE uri

By default is your local machine.

But 
`./ephimeral_sharing -uri=qemu+tcp://remote.com/system`

will use other uri


This will run with 1000 goroutines. (default)


# Customisation:

feel free to experiment.

This project use the go module support so make sure if you build it , to follow the golang modules convention 
