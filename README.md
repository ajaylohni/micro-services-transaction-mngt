# micro-services-transaction-mngt

![](Images/Project.jpg)

To Run this You have to build coustome Image using dockerfile which is at root of the repo. To build 
```sh
$ docker build -t golang:postgres . 
```
It will download extra packages like gorilla mux or lib/pq.
This docker image will be the base image for all other services.

after that run
```sh
$ docker-compose build
$ docker-compose up
```
or
```sh
$docker-compose up --build
```
To see the opening ports
```sh
$docker ps
```


