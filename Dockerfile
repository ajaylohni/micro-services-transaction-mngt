FROM golang:1.10
RUN go get -u github.com/streadway/amqp
RUN go get -u github.com/go-sql-driver/mysql
RUN go get -u github.com/patrickmn/go-cache
RUN go get -u github.com/rs/xid
RUN go get -u github.com/lib/pq
RUN go get -u github.com/gorilla/mux
RUN go get -u github.com/gorilla/handlers