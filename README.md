# micro-services-transaction-mngt
- To build custom image for services 
```sh
$ docker built -t golang:custom .
```
- To up all sercives 
```sh
$ docker-compose up --build
```
- To execute order service
```sh
$ docker exec -it Order ./orderService
```
