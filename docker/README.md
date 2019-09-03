# Docker

This directory contains the build script and files to create and run the SSHCA service as a docker container.  
The container created will mount a persistent container in which the certificates will be created.  

The following commands are available through the make file:

`make generate`  
This command is the first command you aught to run, it builds and creates the docker image and the certificates 
used by the service.

`make serve`  
The serve command starts the service as a docker container.

`make clean`  
Clean will remove all running or dormant containers which are using the volume that the certificates are stored in,
after removal, it will delete the volume so that a new one can be created.  

_Observe: By running the clean command, all certificates will be deleted and you will not be able to create new ones._
