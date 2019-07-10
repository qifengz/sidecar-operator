# step1： 根据服务生成KEY和CSR文件
# domain name 在k8s中一般为$(service_name).$(namespace).svc(也支持通配符)
mkdir /tmp/cert
csr_file=/tmp/cert/csr.pem
key_file=/tmp/cert/key.pem

domain_name="webhook-server-service.operator-system.svc"

openssl req -out "$csr_file" -new -newkey rsa:2048 -nodes -keyout "$key_file" -subj /C=un/ST=st/L=l/O=o/OU=ou/CN="$domain_name"


# step 2: 把生成的待签发文件（csr）
root_ca_crt_file=/tmp/cert/ca-cert.pem
root_ca_key_file=/tmp/cert/ca-key.pem
crt_file=/tmp/cert/cert.pem

openssl genrsa -out "$root_ca_key_file" 2048
openssl req -x509 -new -nodes -key "$root_ca_key_file" -days 10000 -out "$root_ca_crt_file" -subj /C=un/ST=st/L=l/O=o/OU=ou/CN="$domain_name"

openssl x509 -req -days 3650 -in "$csr_file" -CA "$root_ca_crt_file" -CAkey "$root_ca_key_file" -CAcreateserial -out "$crt_file"
