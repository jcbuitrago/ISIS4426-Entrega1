#!/bin/bash

export PGHOST=$(RDS_ENDPOINT)
export PGPORT="5432"
export PGUSER=$(aws ssm get-parameter --name /anb/db/user --with-decryption --query 'Parameter.Value' --output text)
export PGPASSWORD=$(aws ssm get-parameter --name /anb/db/password --with-decryption --query 'Parameter.Value' --output text)
export PGDATABASE="postgres"   # e.g., appdb
export PGSSLMODE="require"             # recommended for RDS

psql -v ON_ERROR_STOP=1 -f /home/ubuntu/ISIS4426-Entrega1/db/init.sql

