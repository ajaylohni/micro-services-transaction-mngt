FROM golang:custom
RUN mkdir -p Order
WORKDIR /Order
ADD . /Order
# EXPOSE 9001
RUN go build ./orderService.go
CMD [ "./orderService" ]