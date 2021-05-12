package RabbitMQ

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"grpcdemo/rabbitmqdemo/models"
	"log"
)

//定义一个连接 url 格式 amqp://账号:密码@rabbitmq服务器地址:端口号/vhost
const MQURL = "amqp://penguiner:penguiner@127.0.0.1:5672/penguin"

type RabbitMQ struct {
	conn *amqp.Connection
	channel *amqp.Channel
	//队列名称
	QueueName string
	//交换机
	Exchange string
	//key
	Key string
	//连接信息
	Mqurl string
}


//创建结构体实例RabbitMQ
func NewRabbitMQ(queueName , exchange ,key string ) *RabbitMQ  {
	rabbitmq:= &RabbitMQ{
		QueueName: queueName,
		Exchange:  exchange,
		Key:       key,
		Mqurl: MQURL,
	}
	var err error
	//创建rabbitmq 连接
	rabbitmq.conn,err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err,"创建连接错误")
	rabbitmq.channel,err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err,"获取channel失败")
	return rabbitmq
}


//断开channel和connection
func (r *RabbitMQ) Destroy()  {
	r.channel.Close()
	r.conn.Close()
}


//错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s",message,err)
		panic(fmt.Sprintf("%s:%s",message,err))
	}
}

//创建简单RabbitMQ实例
func NewRabbitMQSimple(queueName string)*RabbitMQ  {
	return NewRabbitMQ(queueName,"","")

}

func (r *RabbitMQ)PublicSimple(message string)  {
	//1.申请队列,如果队列不存在会自动创建，如果存在则跳过创建
	//保证队列存在，消息能发送到队列中
	_,err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞
		false,
		//额外属性
		nil,


		)
	if err != nil {
		fmt.Println(err)
	}
	//2.发送消息到队列中
	r.channel.Publish(
		r.Exchange,
		r.QueueName,
		//如果为true，根据exchange 类型和routkey 规则，如果无法找到符合条件的队列那么会把消息
		//发送的消息返回给发送者
		false,
		//如果为true，当exchange发送消息到队列后发现队列上没有绑定消费者，则会把消息发还给发送者
		false,
		amqp.Publishing{
			ContentType:     "text/plain",
			Body:            []byte(message),
		})

}

func (r *RabbitMQ) ConsumeSimple(messages models.MessageInterface) {
	//1.申请队列,如果队列不存在会自动创建，如果存在则跳过创建
	//保证队列存在，消息能发送到队列中
	_,err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞
		false,
		//额外属性
		nil,


	)
	if err != nil {
		fmt.Println(err)
	}
	//消费者流控
	//三个参数 1.当前消费者一次能接受最大消息数量 2.服务器传递的最大容量 3.如果设置true 对channel可用
	r.channel.Qos(1,0,false)


	//2.接受消息
	msgs,err := r.channel.Consume(
		r.QueueName,
		//用来区分多个消费者
		"",
		//是否自动应答
		false,
		//是否具有排他性
		false,
		//如果设置为true,表示不能将同一个connection中的消息传递给这个connection中的消费者
		false,
		//队列消费是否阻塞
		false,
		nil,

	)
	if err != nil {
		fmt.Println(err)
	}

	forever := make(chan bool)
	//启用协程处理消息
	go func() {
		for d:=range msgs{
			//实现我们要处理的逻辑函数
			log.Printf("Received a message : %s",d.Body)
			message := &models.Message{}
			err:=json.Unmarshal([]byte(d.Body),message)
			if err != nil {
				fmt.Println(err)
			}
			messages.MessagesAdd(message)

			//如果为true 表示确认所有未确认的消息
			//为false表示确认当前消息
			d.Ack(false)
		}
	}()

	log.Printf("[*] waiting for messages,to exit press CTRL+C")
	<-forever
}

//订阅模式创建RabbitMQ实例
func NewRabbitMQPubSub(exchange string) *RabbitMQ {
	rabbitmq:= NewRabbitMQ("",exchange,"")
	var err error
	//创建rabbitmq 连接
	rabbitmq.conn,err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err,"创建连接错误")
	rabbitmq.channel,err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err,"获取channel失败")
	return rabbitmq
}

func (r *RabbitMQ) PublishPub(message string) {
	//1.尝试创建交换机
	err:= r.channel.ExchangeDeclare(
		r.Exchange,
		//交换机的类型为广播
		"fanout",
		true,
		false,
		//true表示这个exchange不可以被ckient用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		false,
		nil,

		)
	r.failOnErr(err,"Failed to declare an exchange")

	//发送消息
	err= r.channel.Publish(r.Exchange,"",false,false,amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(message),
	})
}

func (r *RabbitMQ) RecieveSub() {
	//1.尝试创建交换机
	err:= r.channel.ExchangeDeclare(
		r.Exchange,
		//交换机的类型为广播
		"fanout",
		true,
		false,
		//true表示这个exchange不可以被ckient用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		false,
		nil,

	)
	r.failOnErr(err,"Failed to declare an exchange")

	//2.试探性创建队列，注意队列不要写名称
	q,err:=r.channel.QueueDeclare(
		"",//随机生产队列名称
		false,
		false,
		true,
		false,
		nil,
		)
	r.failOnErr(err,"failed to declare a queue")

	//绑定队列到exchange 中
	err=r.channel.QueueBind(
		q.Name,
		//在pub/sub模式下，这里的key要为空
		"",
		r.Exchange,
		false,
		nil,

		)
	//消费消息
	message,err:=r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
		)
	forever:=make(chan bool)

	go func() {
		for d:= range message{
			log.Printf("received a message : %s",d.Body)
		}
	}()

	fmt.Println("退出请按 CTRL+C \n")
	<-forever
}

//路由模式
//创建RabbitMQ实例
func NewRabbitMQRouting(exchange string,routingKey string) *RabbitMQ {
	rabbitmq:= NewRabbitMQ("",exchange,routingKey)
	var err error
	//创建rabbitmq 连接
	rabbitmq.conn,err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err,"创建连接错误")
	rabbitmq.channel,err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err,"获取channel失败")
	return rabbitmq
}

//路由模式发送消息
func (r *RabbitMQ) PublishRouting(message string) {
	//1.尝试创建交换机
	err:= r.channel.ExchangeDeclare(
		r.Exchange,
		//交换机的类型为广播
		"direct",
		true,
		false,
		//true表示这个exchange不可以被ckient用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		false,
		nil,

	)
	r.failOnErr(err,"Failed to declare an exchange")

	//发送消息
	err= r.channel.Publish(
		r.Exchange,
		//设置
		r.Key,
		false,
		false,
		amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(message),
	})
}

func (r *RabbitMQ) RecieveRouting() {
	//1.尝试创建交换机
	err:= r.channel.ExchangeDeclare(
		r.Exchange,
		//交换机的类型为广播
		"direct",
		true,
		false,
		//true表示这个exchange不可以被ckient用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		false,
		nil,

	)
	r.failOnErr(err,"Failed to declare an exchange")

	//2.试探性创建队列，注意队列不要写名称
	q,err:=r.channel.QueueDeclare(
		"",//随机生产队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err,"failed to declare a queue")

	//绑定队列到exchange 中
	err=r.channel.QueueBind(
		q.Name,
		//在pub/sub模式下，这里的key要为空
		r.Key,
		r.Exchange,
		false,
		nil,

	)
	//消费消息
	message,err:=r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	forever:=make(chan bool)

	go func() {
		for d:= range message{
			log.Printf("received a message : %s",d.Body)
		}
	}()

	fmt.Println("退出请按 CTRL+C \n")
	<-forever
}
