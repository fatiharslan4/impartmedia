#!/bin/bash

#[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
conn="mysql://impart_super:supersecretpassword@tcp(localhost:3306)/impart?tls=skip-verify&autocommit=true"
migrate -database "${conn}" -path ./schemas/migrations "${1}"

## Dev db
#conn="mysql://impart_db_admin:1pjj82aRrkyFMYnmUZgRfBdLrhb1pjj7gqIJe@tcp(localhost:3306)/impart?tls=skip-verify&autocommit=true"
#migrate -database "${conn}" -path ./schemas/migrations "${1}"
