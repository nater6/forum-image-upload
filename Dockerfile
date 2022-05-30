FROM golang:1.17.3

RUN mkdir /docker-practice

ADD . /docker-practice


WORKDIR /docker-practice
RUN go mod tidy
RUN go build -o main .

CMD ["/docker-practice/main"]