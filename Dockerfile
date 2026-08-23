FROM golang:1.26.1-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -o /out/nanjing-tourism-hub ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/nanjing-tourism-hub /usr/local/bin/nanjing-tourism-hub
RUN mkdir -p /data && chown app:app /data
USER app
ENV TOURISM_HTTP_ADDR=:52822 TOURISM_DATABASE_PATH=/data/nanjing.db
EXPOSE 52822
ENTRYPOINT ["/usr/local/bin/nanjing-tourism-hub"]

