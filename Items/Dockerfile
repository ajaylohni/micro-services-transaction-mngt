FROM golang:custom
RUN mkdir -p Item
WORKDIR /Item
ADD . /Item
EXPOSE 9002
RUN go build ./itemService.go
CMD [ "./itemService" ]