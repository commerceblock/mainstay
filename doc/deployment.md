# Mainstay service deployment

## Instance

The Mainstay service requires a regular cloud compute instance with > 2 GB memory and > 50 GB SSD, with Ubuntu v20.04 system. 

## Installation

### Clone mainstay server:

`git clone https://github.com/commerceblock/mainstay.git`

### Clone mainstay-mvc:

`git clone https://github.com/commerceblock/mainstay.git`

### Install mongodb:

`sudo apt update`

`sudo apt install mongodb-org`

`sudo systemctl start mongod.service`

`sudo systemctl enable mongod`

### Install s3cmd

`apt install s3cmd`

Configure s3cmd:

`s3cmd --configure`

Enter values/credentials from the Vultr object storage `backup` when prompted. 

(https://www.vultr.com/docs/how-to-use-s3cmd-with-vultr-object-storage/)

### Retrieve and restore database

Get mongodb dump file:

`s3cmd get s3://mercury-db/mainstay/dump.tar.gz`

Unpack:

`tar -zxvf dump.tar.gz`

Restore DB:

`mongorestore`

### Build mainstay server:

Install go:

`wget https://dl.google.com/go/go1.19.4.linux-amd64.tar.gz`
`tar -C /usr/local -xzf go1.19.4.linux-amd64.tar.gz`
`export PATH=$PATH:/usr/local/go/bin`

```
go env GOROOT GOPATH

export GOPATH=`go env GOPATH`

mkdir $GOPATH/src
mkdir $GOPATH/bin

export GOBIN=$GOPATH/bin

git clone https://github.com/commerceblock/mainstay $GOPATH/src/mainstay
cd $GOPATH/src/mainstay
s3cmd get s3://mercury-db/mainstay/mainstay

PATH="$GOPATH/bin:$PATH"

go get
```

Set txsigner config:

Edit `cmd/txsigningtool/conf.json`

with tx signer config on Lastpass. 

Edit `config/conf.json`

with mainstay config on Lastpass. 

### Launch mainstay

`./mainstay > mainstay.log &`
`disown`

Run signer - enter command in 'Mainstay keys' in Lastpass. 

Then: `disown`

### Run MVC backend

```
cd
cd mainstay-mvc
```

Run command in Lastpass: Mainstay MVC

### Run frontend

`s3cmd get s3://mercury-db/mainstay/cert.pem`
`s3cmd get s3://mercury-db/mainstay/key.pem`

`HOST_API="127.0.0.1" PORT_API="4000" PORT="80"  webpack-dev-server --https --cert ./cert.pem --key ./key.pem`
