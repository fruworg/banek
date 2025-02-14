# build
FROM golang:alpine AS build
WORKDIR /app
COPY main.go ./
RUN go mod init main
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o banek main.go

# deploy
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/banek .
EXPOSE 9999
CMD ["/app/banek", "--content", "./config/content.json", "--html", "./config/template.html"]
