FROM golang:custom
RUN mkdir -p Payment
WORKDIR /Payment
ADD . /Payment
EXPOSE 9003
RUN go build ./paymentService.go
CMD [ "./paymentService" ]