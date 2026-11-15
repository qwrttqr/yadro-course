FROM golang:1.27 AS build

COPY . /src
RUN cd /src && CGO_ENABLED=0 go build -o /fileserver /src/cmd/api/server.go

FROM alpine:3.20

COPY --from=build /fileserver /fileserver

ENTRYPOINT [ "/fileserver" ]
