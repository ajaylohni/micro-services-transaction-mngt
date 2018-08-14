FROM golang:postgres
RUN mkdir -p PushMsg
WORKDIR /PushMsg
ADD . /PushMsg
EXPOSE 8006
RUN go build ./pushMsg.go
CMD [ "./pushMsg" ]
