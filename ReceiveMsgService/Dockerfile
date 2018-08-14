FROM golang:postgres
RUN mkdir -p ReceiveMsg
WORKDIR /ReceiveMsg
ADD . /ReceiveMsg
RUN go build ./receiveService.go
CMD [ "./receiveService" ]
