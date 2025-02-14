# build
FROM golang:latest AS build
WORKDIR /app
COPY main.go ./
RUN go mod init main
RUN go mod tidy
RUN go build -o banek

# deploy
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/banek .
EXPOSE 9999
CMD ["/app/banek"]
