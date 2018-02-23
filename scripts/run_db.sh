sudo docker run --rm \
-e "POSTGRES_PASSWORD=badgerboi" \
-e "POSTGRES_USER="badgerboi \
-e "POSTGRES_DB=badgerboi" \
--network=badgerboinet \
--name=db \
postgres \
-c 'max_connections=200'
