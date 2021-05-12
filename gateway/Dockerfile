FROM ubuntu

WORKDIR /app

COPY ./_output/gateway /app/gateway
COPY ./tls /app/tls

ENV GO_ENV="development"

CMD ["/app/gateway"]