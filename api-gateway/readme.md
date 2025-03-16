

docker run -d -p 6831:6831/udp -p 14268:14268 -p 16686:16686 jaegertracing/all-in-one:latest

jaegertracing: http://localhost:16686/





docker run -d -p 9090:9090 --network host -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus

now check this and status is UP http://localhost:9090/targets


Prometheus UI: http://localhost:9090

curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"email": "user@example.com", "password": "password123"}'


মেট্রিক্স দেখো: http://localhost:8080/metrics
ট্রেসিং দেখো: Jaeger UI-তে।


লগিং: zap দিয়ে স্ট্রাকচার্ড লগ কনসোলে দেখা যাবে।
ট্রেসিং: Jaeger-এ প্রতিটি রিকোয়েস্টের ট্রেস দেখা যাবে।
মনিটরিং: Prometheus-এ রিকোয়েস্ট কাউন্ট এবং লেটেন্সি মেট্রিক্স পাবে।


docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
5672: AMQP পোর্ট (মেসেজিং-এর জন্য)।
15672: Management UI পোর্ট।
Management UI দেখতে: http://localhost:15672 (ডিফল্ট username: guest, password: guest)।

