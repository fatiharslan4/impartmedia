FROM golang:1.15-buster as build
WORKDIR /app
COPY . /app
RUN go build -mod=vendor -o /app/impart-backend /app/cmd/main.go


FROM debian:buster-slim
RUN apt-get -y update && \
  apt-get install -y curl wget ca-certificates
WORKDIR /app
COPY --from=build /app/impart-backend /app/

ENTRYPOINT ["/app/impart-backend"]