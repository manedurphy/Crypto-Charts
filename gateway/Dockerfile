FROM ubuntu

WORKDIR /app

COPY ./_output/gateway /app/gateway

ENV GO_ENV="development"

CMD ["/app/gateway"]