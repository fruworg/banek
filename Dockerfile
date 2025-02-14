# build
FROM golang:latest as builder
WORKDIR /app
COPY main.go ./
RUN go mod init main
RUN go mod tidy
RUN go build -o banek

# deploy
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/banek .
EXPOSE 9999
CMD ["/app/bin/banek"]
