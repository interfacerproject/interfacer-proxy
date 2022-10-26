FROM golang:1.19-alpine3.16 AS builder
ADD . /app
WORKDIR /app
RUN ENABLE_CGO=0 go build -o interfacer-gateway

FROM alpine:3.16
WORKDIR /root
ENV PORT=80
EXPOSE 80
COPY --from=builder /app/interfacer-gateway /root/
CMD ["/root/interfacer-gateway"]
