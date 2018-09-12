FROM golang:1.11 as builder
WORKDIR /go/src/github.com/yuriichv/demo-discount-service/
COPY main.go .
RUN go get -d -v github.com/openzipkin/zipkin-go/middleware/http golang.org/x/net/html github.com/openzipkin/zipkin-go/model github.com/openzipkin/zipkin-go/reporter/http github.com/openzipkin/zipkin-go github.com/gorilla/mux\
  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY --from=builder /go/src/github.com/yuriichv/demo-discount-service/app .
CMD ["./app"]  
ENTRYPOINT [""]
