FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=America/Mexico_City

WORKDIR /app
COPY server .
RUN chmod +x ./server

EXPOSE 8080

CMD ["./server"]
