# picuploader

This programs, writen in GO, monitors a folded and uploads any file placed
there to an HTTP server. 

It has been tested to monitor and upload picture files to a locally installed
NextCloud server in my network. However, it can be used to upload files to any
HTTP server, since it uses curl to upload the files.

Originally written for Linux, the program can easily be converted to Windows
due to GO easy of portability.

All the parameters will be read from a config file, an example of which is in 
the package.

Once started, picuploader will run continously until it's terminated with CRTL-C,
sleeping at intervals predefined in the config file.

Essentially, here's what you need to use this program:
<br>
. File server (usually HTTP) configured to receive files (take a look at https://nextcloud.com)
<br>
. curl installed in the same machine as the picuploader
<br>
. picservice.conf file in your home folder
<br>
. picuploader


To build the program:
<br>
. first, download the source from here
<br>
. open a terminal window on the folder you downloaded the source into
<br>
. issue this command:
<br>
  go build service.go
<br>
. to execute the standalone app, issue this command
<br>
  ./service   


That's it. Enjoy!

