// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// MONGO DB INIT SCRIPT
// Script initialises all collections currently used
// by mainstay service and mainstay website
//
// Also set up two different roles; one for Api, one for Service
//
// Script assumes an admin user with user:pass set in DB_USER/DB_PASS env
// Make sure mongo db running in auth mode:
// mongod -auth
// Run this using:
// mongo --eval "var db_host='$DB_HOST'; db_name='$DB_NAME_MAINSTAY'; var db_user='$DB_USER'; var db_pass ='$DB_PASS'" scripts/db-init.js

db = connect(db_user + ":" + db_pass + "@" + db_host + "/admin");

// Connect/create mainstayX database
db = db.getSiblingDB(db_name)

// Create collections
print("creating collections")
db.createCollection("Attestation")
db.createCollection("AttestationInfo")
db.createCollection("ClientCommitment")
db.createCollection("ClientDetails")
db.createCollection("MerkleCommitment")
db.createCollection("MerkleProof")
print(db.getCollectionNames())

// Create roles
print("creating roles")
db.dropRole("mainstayApi")
db.dropRole("mainstayService")

// mainstayApi role
// This allows only writing to client collections
// and reading from all the other collections
db.createRole(
{
    role: "mainstayApi",
    privileges: [
        { resource: { db: db_name, collection: "Attestation" }, actions: [ "find"] },
        { resource: { db: db_name, collection: "AttestationInfo" }, actions: [ "find"] },
        { resource: { db: db_name, collection: "MerkleCommitment" }, actions: [ "find"] },
        { resource: { db: db_name, collection: "MerkleProof" }, actions: [ "find"] },
        { resource: { db: db_name, collection: "ClientCommitment" }, actions: [ "find", "update", "insert"] },
        { resource: { db: db_name, collection: "ClientDetails" }, actions: [ "find", "update", "insert"] },
    ],
    roles: []
}
)

// mainstayService role
// This allows writing to all collections except
// ClientCommitment/ClientDetails which only API is allowed to write to
db.createRole(
{
    role: "mainstayService",
    privileges: [
        { resource: { db: db_name, collection: "Attestation" }, actions: ["find", "update", "insert"] },
        { resource: { db: db_name, collection: "AttestationInfo" }, actions: ["find", "update", "insert"] },
        { resource: { db: db_name, collection: "MerkleCommitment" }, actions: ["find", "update", "insert"] },
        { resource: { db: db_name, collection: "MerkleProof" }, actions: ["find", "update", "insert"] },
        { resource: { db: db_name, collection: "ClientCommitment" }, actions: ["find"] },
        { resource: { db: db_name, collection: "ClientDetails" }, actions: ["find"] },
    ],
    roles: []
}
)

// Create two users - one for Api one for Service
db.dropUser("apiUser")
db.createUser({user: "apiUser", pwd: "apiPass", roles: ["mainstayApi"]});
db.dropUser("serviceUser")
db.createUser({user: "serviceUser", pwd: "servicePass", roles: ["mainstayService"]});
