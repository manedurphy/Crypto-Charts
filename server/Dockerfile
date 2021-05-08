FROM ubuntu

RUN apt update && apt install ca-certificates -y

WORKDIR /app

COPY ./_output/server /app/server

CMD ["/app/server"]