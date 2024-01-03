# Build stage
FROM golang:latest AS builder

WORKDIR /src
# RUN apt update && apt install git

COPY . .
RUN go mod download


RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./entrypoints/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /src/server /usr/bin/server
EXPOSE 8080
ENTRYPOINT ["/usr/bin/server"]

# CMD ["./server"]


