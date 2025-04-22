#!/bin/sh

openssl ecparam -name prime256v1 -genkey -noout -out CA.key
openssl req -x509 -new -nodes -key CA.key -sha256 -days 365 -out CA.pem -subj "/CN=Root CA"

openssl ecparam -name prime256v1 -genkey -noout -out server.key
openssl req -new -key server.key -out server.csr -config ./server.conf
openssl x509 -req -in server.csr -CA CA.pem -CAkey CA.key -CAcreateserial -out server.crt -days 365 -sha256 -extensions req_ext -extfile ./server.conf

openssl ecparam -name prime256v1 -genkey -noout -out client.key
openssl req -new -key client.key -out client.csr -config ./client.conf
openssl x509 -req -in client.csr -CA CA.pem -CAkey CA.key -CAcreateserial -out client.crt -days 365 -sha256 -extensions req_ext -extfile ./client.conf


openssl ecparam -name prime256v1 -genkey -noout -out CA_fail.key
openssl req -x509 -new -nodes -key CA_fail.key -sha256 -days 365 -out CA_fail.pem -subj "/CN=Root CA" -addext "subjectAltName = DNS:localhost,IP:0.0.0.0"

openssl ecparam -name prime256v1 -genkey -noout -out client_fail.key
openssl req -new -key client_fail.key -out client_fail.csr -subj "/CN=fail_user" -config ./client.conf
openssl x509 -req -in client_fail.csr -CA CA_fail.pem -CAkey CA_fail.key -CAcreateserial -out client_fail.crt -days 365 -sha256 -extensions req_ext -extfile ./client.conf
