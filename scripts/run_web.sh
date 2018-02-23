sudo docker run --rm -e "ROLE=web" -e "QUOTE_SERVER_URL=192.168.1.152:4450" --network=badgerboinet -p 8082:8082 --name=web badgerboi_web
