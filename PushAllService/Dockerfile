FROM golang:postgres
RUN mkdir -p InitPushService
WORKDIR /InitPushService
ADD . /InitPushService
RUN go build ./initPushService.go
CMD [ "./initPushService" ]
