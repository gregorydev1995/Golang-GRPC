package RabbitMQ

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/unknwon/goconfig"
	"gome/api"
	"gome/gomengine/Engine"
	"gome/gomengine/Redis"
	"log"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	Queuename string // 列表名
	Exchange  string // 交换机
	Key       string
	MQurl     string
}

func NewRabbitMq(queuename, exchage, key string) *RabbitMQ {
	config, err := goconfig.LoadConfigFile("config.ini")
	host, _ := config.GetValue("rabbitmq", "host")
	port, _ := config.GetValue("rabbitmq", "port")
	username, _ := config.GetValue("rabbitmq", "username")
	password, _ := config.GetValue("rabbitmq", "password")
	if err != nil {
		panic("配置读取失败")
	}
	// MQURL 格式 amqp://账号:密码@rabbitmq服务器地址:端口号/vhost
	MQurl := "amqp://" + username + ":" + password + "@" + host + ":" + port + "/"

	rabbitmq := &RabbitMQ{
		Queuename: queuename,
		Exchange:  exchage,
		Key:       key,
		MQurl:     MQurl,
	}
	rabbitmq.conn, err = amqp.Dial(rabbitmq.MQurl)
	rabbitmq.failOnErr(err, "连接MQ错误")

	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "获取channel失败")

	return rabbitmq
}

func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

func (r *RabbitMQ) Destory() {
	_ = r.channel.Close()
	_ = r.conn.Close()
}

func NewSimpleRabbitMQ(queuename string) *RabbitMQ {
	return NewRabbitMq(queuename, "", "")
}

func (r *RabbitMQ) PublishSimple(message []byte) {
	//1. 申请队列，如果队列不存在会自动创建，如何存在则跳过创建
	_, err := r.channel.QueueDeclare(
		r.Queuename,
		false, // 是否持久化
		false, // 是否自动删除
		false, // 是否具有排他性？
		false, // 是否阻塞
		nil,   // 其它属性
	)
	if err != nil {
		fmt.Println(err)
	}

	r.channel.Publish(
		r.Exchange,
		r.Queuename,
		false, // 如果为true, 会根据exchange类型和routkey规则，如果无法找到符合条件的队列那么会把发送的消息返回给发送者
		false, // 如果为true, 当exchange发送消息到队列后发现队列上没有绑定消费者，则会把消息发还给发送者
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
}

func (r *RabbitMQ) ConsumeSimple() {
	_, err := r.channel.QueueDeclare(
		r.Queuename,
		false, // 是否持久化
		false, // 是否自动删除
		false, // 是否具有排他性？
		false, // 是否阻塞
		nil,   // 其它属性
	)
	if err != nil {
		fmt.Println(err)
	}

	msgs, err := r.channel.Consume(
		r.Queuename,
		"",    // 用来区分多个消费者
		true,  // 是否自动应答
		false, // 是否具有排他性
		false, // 如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false, // 队列消费是否阻塞
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	gomerdb := Redis.NewRedisClient()
	forever := make(chan bool)
	// 启用协程处理消息
	//go func() {
	for d := range msgs {
		//log.Printf("Received a message: %s", d.Body)
		//fmt.Println(d.Body)
		order := api.OrderRequest{}
		err := json.Unmarshal(d.Body, &order)
		if err != nil {
			fmt.Println(err)
		}
		node := Engine.NewOrderNode(order)
		fmt.Println("-------------", node)
		var ctx = context.Background()
		err = gomerdb.Set(ctx, "testte", order.Symbol, 0).Err()
		//if err != nil {
		//panic(err)
		//}
		fmt.Println(order.Oid)
		fmt.Println(order.Symbol)
	}

	//}()
	log.Printf("[*] Waiting for message, To exit press CTRL+C")
	<-forever
}
