#!/bin/sh

mkdir ca client server

## Root CA first
echo Creating root
cd ca

mkdir certs crl newcerts private

touch index.txt
echo 1000 > serial

#Generate key
openssl genrsa -aes256 -out private/key.pem 2048

#Create root certificate
openssl req -config ../openssl.cnf \
      -key private/key.pem \
      -new -x509 -days 7300 -sha256 -extensions v3_intermediate_ca \
      -out certs/cert.pem

#Verify cert
openssl x509 -noout -text -in certs/cert.pem

tree .


#Create server cert
echo Creating server
cd ../server

# Create key
openssl genrsa -out key.pem 2048

#Create cert
openssl req -config ../openssl.cnf \
      -key key.pem \
      -new -sha256 -out csr.pem

#Sign cert
openssl ca -config ../openssl.cnf \
      -extensions server_cert -days 375 -notext -md sha256 \
      -in csr.pem \
      -out cert.pem

cat ../ca/index.txt


### Create client
echo Creating client
cd ../client

# Create key
openssl genrsa -out key.pem 2048

#Create cert
openssl req -config ../openssl.cnf \
      -key key.pem \
      -new -sha256 -out csr.pem

#Sign cert
openssl ca -config ../openssl.cnf \
      -extensions usr_cert -days 375 -notext -md sha256 \
      -in csr.pem \
      -out cert.pem

#Create export pk12
#This key needs to get imported by the os's key store
openssl pkcs12 -export -inkey key.pem -in cert.pem -out client.p12

cat ../ca/index.txt

