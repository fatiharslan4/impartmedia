#!/bin/bash

echo "Generating DB code..."
sqlboiler mysql --add-soft-deletes
echo "Done."
#go test ./pkg/models/dbmodels/...