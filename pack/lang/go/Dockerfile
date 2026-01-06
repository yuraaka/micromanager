FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o /tmp/{{.ServiceBinaryName}} ./services/{{.ServiceDir}}/server

FROM alpine:3.22
EXPOSE 8000
ENV PORT=8000
WORKDIR /app
COPY --from=build /tmp/{{.ServiceBinaryName}} /app/{{.ServiceBinaryName}}
ENTRYPOINT ["/app/{{.ServiceBinaryName}}"]
