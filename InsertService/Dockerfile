FROM golang:postgres
RUN mkdir -p InsertService
WORKDIR /InsertService
ADD . /InsertService
RUN go build ./insertService.go
CMD [ "./insertService" ]
