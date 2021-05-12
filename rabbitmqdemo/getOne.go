package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"grpcdemo/rabbitmqdemo/models"
	RabbitMQ "grpcdemo/rabbitmqdemo/rabbitmq"
	"net/http"
	"sync"
)

var sum int64 =0
//预存商品数量
var productNum int64=10000

//互斥锁
var mutex sync.Mutex


//获取秒杀商品
func GetOneProduct() bool {
	mutex.Lock()
	defer mutex.Unlock()

	//判断是否超限
	if sum < productNum{
		sum+=1
		return true
	}
	return false


}



func GetProduct(c *gin.Context)  {
	//if GetOneProduct() {
	//	c.String(200,"true")
	//	return
	//}



	message := models.NewMessage(1,1)
	byteMessage,err :=json.Marshal(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError,err)
		return
	}

	rabbitmq := RabbitMQ.NewRabbitMQSimple("penguin")
	rabbitmq.PublicSimple(string(byteMessage))
	c.String(200,"true")

}



func main() {
	r:=gin.Default()


	r.GET("/getOne",GetProduct)
	
	
	
	r.Run(":8080")
}