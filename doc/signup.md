## Sign up process

Prospective clients will need to provide details required for KYC as well as an ECDSA public key (either compressed or uncompressed). The corresponding private key will be used by the client to sign commitments before sending to Mainstay API.

These should be shared with Commerceblock via:

- Mainstay website (under development)
- Manually by communicating with the CommerceBlock team

To signup a client the `clientsignuptool` can be used. The tool requires appropriate connectivity and permissions to the Mainstay database. Example use (providing only public key and Client Name):

```
*********************************************
************ Client Signup Tool *************
*********************************************

existing clients
client_position: 0 pubkey: 03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33 name: My First Client
client_position: 1 pubkey: 03c4fab0ba2ae849d95cfce69ee483860f058a3d9858475722b22d88143c808684 name: Nikos
client_position: 2 pubkey: 03d3825563181ef60f869a95415348285e70feb8259d01ed0e468d969ed9ff74b9 name: Ocean

next available position: 3

*********************************************
************ Client Pubkey Info *************
*********************************************

Insert pubkey: 021805882d0939594b34f1c31e8d6d9cc19700c6e08cc9d83c267ae318ec52b796
pubkey verified

*********************************************
***** Client Auth Token identification ******
*********************************************

new-uuid: e0950c03-2f71-42b8-af5c-a564b26a9f97

*********************************************
*********** Inserting New Client ************
*********************************************

Insert client name: Test Client
NEW CLIENT DETAILS
client_position: 3
auth_token: e0950c03-2f71-42b8-af5c-a564b26a9f97
pubkey: 021805882d0939594b34f1c31e8d6d9cc19700c6e08cc9d83c267ae318ec52b796

existing clients
client_position: 0 pubkey: 03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33 name: My First Client
client_position: 1 pubkey: 03c4fab0ba2ae849d95cfce69ee483860f058a3d9858475722b22d88143c808684 name: Nikos
client_position: 2 pubkey: 03d3825563181ef60f869a95415348285e70feb8259d01ed0e468d969ed9ff74b9 name: Ocean
client_position: 3 pubkey: 021805882d0939594b34f1c31e8d6d9cc19700c6e08cc9d83c267ae318ec52b796 name: Test Client

```

The tool will generate a `client_position` and `auth_token` both of which should be sent to the client and used when sending commitments.
