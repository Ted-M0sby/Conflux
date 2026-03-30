FROM golang:1.21-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN go mod tidy && CGO_ENABLED=0 go build -o /mock .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=build /mock /usr/local/bin/mock-backend
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mock-backend"]
