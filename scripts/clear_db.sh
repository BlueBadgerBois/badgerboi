#!/bin/bash
docker cp reset_db.sql db:reset_db.sql \
  && docker exec -it db psql -U badgerboi -a -f reset_db.sql
