FROM golang:postgres
RUN mkdir -p UpdateService
WORKDIR /UpdateService
ADD . /UpdateService
RUN go build ./updateService.go
CMD [ "./updateService" ]
