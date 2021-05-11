SERVER_CN=localhost

openssl genrsa -out ca.key 4096

openssl req -x509 -new -nodes -key ca.key -sha256 -days 1825 -out ca.crt -subj "/CN=${SERVER_CN}"

openssl genrsa -out server.key 4096

openssl req -new -sha256 -key server.key -out server.csr -subj "/CN=${SERVER_CN}"

# openssl req -in server.csr -noout -text

openssl x509 -req -extfile cert.conf -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 1825 -sha256 