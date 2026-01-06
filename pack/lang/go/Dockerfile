FROM golang:1.25.2-alpine3.22 AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /tmp/{{snake .ServiceName}} ./{{snake .ServiceName}}/server

FROM alpine:3.22
EXPOSE 8000
ENV PORT=8000
RUN apk add --no-cache tini
USER nobody
WORKDIR /app
COPY --from=build /tmp/{{snake .ServiceName}} /app/{{snake .ServiceName}}
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./{{snake .ServiceName}}"]
