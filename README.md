Based on article : https://medium.com/@yusufs/getting-started-with-kafka-in-golang-14ccab5fa26
### How To Run
1. Run Kafka Cluster by execute **docker-compose.yml** file by **MY_IP=your-ip docker-compose up**. Find your IP using **ip a** command.
2. For Example, create topic **foo** with 4 partitions and replication factors = 2 using command below.
```
docker run --net=host --rm confluentinc/cp-kafka:5.0.0 kafka-topics --create --topic foo --partitions 4 --replication-factor 2 --if-not-exists --zookeeper localhost:32181
```
3. For show topic use this command below 
```
docker run --net=host --rm confluentinc/cp-kafka:5.0.0 kafka-topics --zookeeper localhost:32181 --list
```
4. To ensure kafka cluster was running, you can use **kafkacat** as producer and consumer
    * Producer example, produce message to **foo** topic and partition 0 : **echo 'publish to partition 0' | kafkacat -P -b localhost:19092,localhost:29092,localhost:39092 -t foo -p 0**
    * Consumer example, consume message from **foo** topic and partition 0 : **kafkacat -C -b localhost:19092,localhost:29092,localhost:39092 -t foo -p 0**
