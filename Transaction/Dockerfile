FROM golang:custom
RUN mkdir -p Transaction
WORKDIR /Transaction
ADD . /Transaction
EXPOSE 9004
RUN go build ./transactionService.go
CMD [ "./transactionService" ]