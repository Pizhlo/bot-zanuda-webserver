FROM golang:alpine AS builder
WORKDIR /app
ADD go.mod .
COPY . .
RUN go build -o server .

FROM ubuntu
WORKDIR /app
COPY --from=builder /app/server /app/server
# copy the ca-certificate.crt from the build stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["./server"]