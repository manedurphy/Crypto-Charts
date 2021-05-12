FROM ubuntu

RUN apt update && apt install ca-certificates -y

WORKDIR /app

COPY ./_output/server /app/server
COPY ./tls /app/tls

CMD ["/app/server"]