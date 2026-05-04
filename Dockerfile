# Build stage
FROM golang:1.26-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server .

# Runtime stage - minimal image
FROM alpine:latest
WORKDIR /app

COPY --from=build /app/server .

EXPOSE 8080

CMD ["./server"]
