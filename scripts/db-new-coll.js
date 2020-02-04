// Copyright (c) 2018 CommerceBlock Team
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// MONGO DB ADD COLLECTION script
// Script that adds a new collection to the db
// and modifies privileges for existing roles
//

db = connect(db_user + ":" + db_pass + "@" + db_host + "/admin");

db = db.getSiblingDB(db_name)

db.createCollection("ClientSignup")

db.grantPrivilegesToRole(
    "mainstayApi",
    [
        { resource: { db: db_name, collection: "ClientSignup" }, actions: [ "find", "update", "insert"] },
    ],
)

db.grantPrivilegesToRole(
    "mainstayService",
    [
        { resource: { db: db_name, collection: "ClientSignup" }, actions: [ "find"] },
    ],
)
