FROM golang:1.22-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=off
COPY go.mod ./
COPY . .
RUN go mod tidy
RUN go build -o server .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8081
CMD ["./server"]
