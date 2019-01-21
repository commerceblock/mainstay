# Config

## Sample Config

```
// Sample config file parsed in config/config.go
// conf_template.json
{
    "staychain": {
        "initTx": "87e56bda501ba6a022f12e178e9f1ac03fb2c07f04e1dfa62ac9e1d83cd840e1",
        "initScript": "51210381324c14a482646e9ad7cf82372021e5ecb9a7e1b67ee168dddf1e97dafe40af210376c091faaeb6bb3b74e0568db5dd499746d99437758a5cb1e60ab38f02e279c352ae",
        "initChaincodes": "0a090f710e47968aee906804f211cf10cde9a11e14908ca0f78cc55dd190ceaa,0a090f710e47968aee906804f211cf10cde9a11e14908ca0f78cc55dd190ceaa",
        "topupAddress": "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB",
        "topupScript": "51210381324c14a482646e9ad7cf92372021e5ecb9a7e1b67ee168dddf1e97dafe40af210376c091faaeb6bb3b74e0568db5dd499746d99437758a5cb1e60ab38f02e279c352ae",
        "regtest": "1"
    },
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

## Config Parameters

### Compulsory

Currently `main` config category is compulsory. This should be made optional in the future as tools that do not require `main` rpc connectivity options use this.

- `main` : configuration options for connection to bitcoin node
    - `rpcurl` : address for rpc connectivity
    - `rpcuser` : user name for rpc connectivity
    - `rpcpass` : password for rpc connectivity
    - `chain`: chain name for inner config, i.e. testnet/regtest/main


The `staychain` category is compulsory and can be set from either .conf file or command line arguments. The configuration below is optional as preferred entry is via command line - [options](#command-line-options).

- `staychain` : configuration options for staychain parameters
    - `initTx` : initial transaction sets the state for the staychain
    - `initScript` : initial script used to derive subsequent staychain addresses
    - `initChaincodes`: chaincodes of init script pubkeys used to derive subsequent staychain addresses
    - `topupAddress` : address to topup the mainstay service
    - `topupScript` : script that requires signing for the topup


Several other subcategories become compulsory only if the base category exists in the `.conf` file.

For the base categories `db` and `signer` the following parameters are compulsory:

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

### Command Line Options

Currently only parameters in the `staychain` category can be parsed through command line arguments.

These command line arguments are:
- `tx` : argument for initTx as above
- `script` : argument for initScript as above
- `chaincodes`: argument for initChaincodes as above
- `addrTopup` : argument for topupAddress as above
- `scriptTopup` : argument for topupScript as above

### Env Variables

All config parameters can be replaced with env variables. An example of this is `config/conf.json`.

The Config struct works by first looking for an env variable with the name set and if an env variable is not found then the config parameter is set to the actual value provided.

If the config argument is not to be used, __no value__ should be set in the conf file. Warnings for invalid argument values are provided in runtime.

### Client Chain Parameters

Parameters used for client chain confirmation tools and are not part of Config struct used by service.

- `clientchain` : configuration options for connectivity to client rpc node

Same configuration options as `main`. The `clientchain` name can be replaced with any name to match the sidechain. See `cmd/confirmationtool/conf.json`. This is not used by Config struct, only by `config::NewClientFromConfig()`.
