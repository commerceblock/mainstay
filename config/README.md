# Config

## Sample Config

```
// Sample config file parsed in config/config.go
// conf_template.json
{
    "main": {
        "rpcurl": "127.0.0.1:18000",
        "rpcuser": "USERNAME",
        "rpcpass": "PASSWORD",
        "chain": "regtest"
    },
    "clientchain": {
        "rpcurl": "127.0.0.1:19000",
        "rpcuser": "USERNAME",
        "rpcpass": "PASSWORD",
        "chain": "main"
    },
    "signer": {
        "publisher": "*:5000",
        "signers": "node0:1000,node1:1001"
    },
    "db": {
        "user":"user",
        "password":"pssword",
        "host":"localhost",
        "port":"27017",
        "name":"mainstay"
    },
    "fees": {
        "minFee": "5",
        "maxFee": "50",
        "feeIncrement": "2"
    },
    "timing": {
        "newAttestationMinutes": "60",
        "handleUnconfirmedMinutes": "60"
    }
}
```

## Config Parameters

### Compulsory

Currently `main` config category is compulsory.

This should be made optional in the future as tools that do not require `main` rpc connectivity options use this Config.

- `main` : configuration options for connection to bitcoin node
    - `rpcurl` : address for rpc connectivity
    - `rpcuser` : user name for rpc connectivity
    - `rpcpass` : password for rpc connectivity
    - `chain`: chain name for inner config, i.e. testnet/regtest/main

Several other subcategories become compulsory only if the base category exists in the `.conf` file.

For the base categories `db` and `signer` the following the following parameters exist:

- `db` : configuration options for database
    - `user` : db user name
    - `password` : db user password
    - `host` : db host address
    - `port` : db host port
    - `name` : db name

- `signer` : zmq signer connectivity options
    - `signers` : list of comma separated addresses (host:port) for connectivity to signers

### Optional

All the remaining conf options are optional. These are explained below:

- `clientchain` : configration options for connectivity to client rpc node

Same configuration options as `main`. The `clientchain` name can be replaced with any name to match the sidechain. See `cmd/confirmationtool/conf.json`.

- `signer`
    - `publisher` : optionally provide host address for main service zmq publisher

Default values are set in `attestation/attestsigner_zmq.go`.

- `fees` : fee configuration parameters for attestation service
    - `minFee` : minimum fee for attestation transactions
    - `maxFee` : maximum fee for attestation transactions
    - `feeIncrement` : fee increment value used when bumping fees

Default values are set in `attestation/attestfees.go`

- `timing` : various timing configuration parameters used by attestation service
    - `newAttestationMinutes` : option in minutes to set frequency of new attestations
    - `handleUnconfirmedMinutes` : option in minutes to set duration of waiting for an unconfirmed transaction before bumping fees

Default values are set in `attestation/attestservice.go`

### Env variables

All config parameters can be replaced with env variables. An example of this is `config/conf.json`.

The Config struct works by first looking for an env variable with the name set and if an env variable is not found then the config parameter is set to the actual value provided.
