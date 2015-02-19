#!/bin/sh

docker rmi gocross 
docker build --tag="gocross" .
docker run  --rm -it -v "$(pwd)":/usr/src/myapp -w /usr/src/myapp gocross

mkdir build
cp ouiserver-linux-amd64 build/ouiserver-linux-amd64
cp Dockerfile-build build/Dockerfile
cd build
docker rmi ouiserver-service
docker build --tag="ouiserver-service" .
docker save ouiserver-service > ../ouiserver-docker-image.tar
cd ..
rm ouiserver-docker-image.tar.bz2
bzip2 ouiserver-docker-image.tar
rm -rf build
