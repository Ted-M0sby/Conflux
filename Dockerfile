FROM golang:1.21-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -o /nexus ./cmd/nexus

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /nexus /app/nexus
COPY configs /app/configs
EXPOSE 8080
ENTRYPOINT ["/app/nexus"]
