sudo docker run \
--rm \
--name=db_proxy \
-e DB_HOST=db \
-e DB_USER=badgerboi \
-e DB_PASSWORD=badgerboi \
--network=badgerboinet \
brainsam/pgbouncer:latest

