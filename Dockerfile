FROM golang

ADD . /go/gohh
WORKDIR /go/gohh

RUN go mod tidy
RUN go build main.go

EXPOSE 8080

ENTRYPOINT ["./main"]
CMD ["--web"]
