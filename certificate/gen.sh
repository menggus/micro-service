rm *.pem

# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -nodes -days 365 -keyout ca-key.pem -out ca-cert.pem -subj "/C=ch/ST=sz/L=sz/O=YR/OU=YR/CN=YR/emailAddress=YR@gmail.com"

echo "CA's self signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# 2.Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=ch/ST=sz/L=sz/O=YR-web/OU=YR-web/CN=YR-web/emailAddress=YR-web@gmail.com"

# 3.Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in server-req.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf

echo "Server's self signed certificate"
openssl x509 -in server-cert.pem -noout -text


# -days 60: expire time
# -nodes: no need password for private