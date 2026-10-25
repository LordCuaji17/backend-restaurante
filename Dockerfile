FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk --no-cache add ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=America/Mexico_City

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
