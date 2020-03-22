# Alcatraz
Alcatraz project implements encrypted file transfer over gRPC.

# Build
To build both the client and server, run:
```
    make build
```
It will create **bin** directory in the root with the server and client.

# How to use
After you build the client and the server, go to **bin** directory.

First, lets run the server:
```
    ./alcatrazd -crt=../certs/localhost.crt -key=../certs/localhost.key -ca=../certs/CertAuth.crt &
```
Certificates flags are mandatory. If you don't specify **-storage** flag, it will use the default which is **"storage"**.

Now, lets run the client:
```
    mkdir upload
    ./alcatraz upload -crt=../certs/Reese.crt -key=../certs/Reese.key -ca=../certs/CertAuth.crt
```
Certificates again are mandatory, however here we have to give as first argument the folder that will be monitored.
If not given, or folder does not exist, client will fail.

Last step is to put a file or folder with files in **upload** folder.
