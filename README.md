Based on article : https://medium.com/@yusufs/getting-started-with-kafka-in-golang-14ccab5fa26
### How To Run
1. Run Kafka Cluster by execute **docker-compose.yml** file by **MY_IP=your-ip docker-compose up**. Find your IP using **ip a** command.
2. After Kafka cluster running, you can create topic **foo** (for example with 4 partitions and replication factors = 2) using command below.
```
docker run --net=host --rm confluentinc/cp-kafka:latest kafka-topics --create --topic foo --partitions 4 --replication-factor 2 --if-not-exists --zookeeper localhost:32181
```
3. For show topic use this command below 
```
docker run --net=host --rm confluentinc/cp-kafka:latest kafka-topics --zookeeper localhost:32181 --list
```
4. To ensure kafka cluster was running, you can use **kafkacat** as producer and consumer
    * Producer example, produce message to **foo** topic and partition 0 : **echo 'publish to partition 0' | kafkacat -P -b localhost:19092,localhost:29092,localhost:39092 -t foo -p 0**
    * Consumer example, consume message from **foo** topic and partition 0 : **kafkacat -C -b localhost:19092,localhost:29092,localhost:39092 -t foo -p 0**

5. Run API for writer by **make run_api**
6. Run API for writer by **make run_api**
7. Make **POST** request to **http://localhost:4500/api/v1/data** with JSON request body below
```
{
	"text": "Your text!"
}
```