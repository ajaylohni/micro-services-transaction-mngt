FROM golang:postgres
RUN mkdir -p DeleteService
WORKDIR /DeleteService
ADD . /DeleteService
RUN go build ./deleteService.go
CMD [ "./deleteService" ]
