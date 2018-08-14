FROM golang:postgres
RUN mkdir -p DisplayService
WORKDIR /DisplayService
ADD . /DisplayService
RUN go build ./displayService.go
CMD [ "./displayService" ]
