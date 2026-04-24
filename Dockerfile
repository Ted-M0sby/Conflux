FROM alpine:3.19 AS deps
RUN apk add --no-cache build-base cmake git

FROM deps AS build
WORKDIR /src
COPY . .
RUN cmake -S . -B build && cmake --build build -j

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /src/build/conflux /app/conflux
COPY configs /app/configs
EXPOSE 8080
ENTRYPOINT ["/app/conflux"]
