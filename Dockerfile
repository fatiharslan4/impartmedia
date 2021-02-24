FROM golang:1.15-buster as build
WORKDIR /app
COPY . /app
RUN go build -mod=vendor -o /app/impart-backend /app/cmd/server/main.go


FROM debian:buster-slim
RUN apt-get -y update && \
  apt-get install -y curl wget ca-certificates
WORKDIR /app
COPY --from=build /app/impart-backend /app/
COPY --from=build /app/schemas/json /app/schemas/json

RUN mkdir -p ~/.aws && \
   echo "[local]" > ~/.aws/config && \
   echo "region=us-east-2" >> ~/.aws/config && \
   echo "output=json" >> ~/.aws/config && \
   echo "[local]" >  ~/.aws/credentials && \
   echo "aws_access_key_id = dummy" >>  ~/.aws/credentials && \
   echo "aws_secret_access_key = dummy" >>  ~/.aws/credentials

RUN cat ~/.aws/config && cat ~/.aws/credentials

ENTRYPOINT ["/app/impart-backend"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=3s --retries=3 \
  CMD curl -f http://localhost/ping || exit 1