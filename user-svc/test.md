docker run -d -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=userdb postgres
docker run -d -p 27017:27017 mongo
docker run -d -p 5672:5672 -p 15672:15672 rabbitmq:3-management
docker run -d -p 6831:6831/udp -p 14268:14268 -p 16686:16686 jaegertracing/all-in-one