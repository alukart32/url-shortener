# Сокращатель URL

Проект по сокращению URL в рамках курса "Продвинутого Go-разраобтчика" от Яндекс.

## Цель проекта

1. Освоить на практике функциональные возможности go.
2. Создать сервис по сокращению URL.

## X.509 Certs

- [create self signed certificate](https://www.golinuxcloud.com/generate-self-signed-certificate-openssl/)
- [client, server certificates](https://www.golinuxcloud.com/openssl-create-client-server-certificate/)

### Openssl create self signed certificate with passphrase

1. Create encrypted password file

    echo secret > mypass
    openssl enc -aes256 -pbkdf2 -salt -in mypass -out mypass.enc

2. Generate private key

    openssl genrsa -des3 -passout file:mypass.enc -out server.key 4096

3. Create Certificate Signing Request

    openssl req -new -key server.key -out server.csr -passin file:mypass.enc

    (To automate this step to create CSR (server.csr) use openssl.cnf
    openssl req -new -key server.key -out server.csr -passin file:mypass.enc -config self_signed_certificate.cnf)


4. Create self signed certificate using openssl x509

    openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt -passin file:mypass.enc

### Openssl create self signed certificate without passphrase

1. Generate private key

    openssl genrsa -out server-noenc.key 4096

2. Create Certificate Signing Request

    openssl req -new -key server-noenc.key -out server-noenc.csr

    (To automate this step to create CSR (server.csr) use openssl.cnf
    openssl req -new -key server-noenc.key -out server-noenc.csr -config self_signed_certificate.cnf)

3. Create self signed certificate using openssl x509

    openssl x509 -req -days 365 -in server-noenc.csr -signkey server-noenc.key -out server-noenc.crt
