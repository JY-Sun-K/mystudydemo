package models

import "fmt"

type Message struct {
	ProductID int64
	UserID	int64
}

func NewMessage(userId int64 ,productId int64) *Message {
	return &Message{
		ProductID: productId,
		UserID:    userId,
	}
}

type MessageInterface interface {
	MessagesAdd(message *Message)
	MessagesShow()
}


type MessageSum struct {
	Messages []*Message
}
func NewMessageSum() MessageInterface {

	return &MessageSum{Messages: make([]*Message,0)}
}

func (m *MessageSum)MessagesAdd(message *Message)  {
	//fmt.Println(message)
	m.Messages=append(m.Messages,message)
	//fmt.Println(m.Messages)
	//for _,me := range m.Messages {
	//	fmt.Println(me.ProductID,"   ",me.UserID)
	//}
	m.MessagesShow()
}
func (m *MessageSum)MessagesShow()  {
	for _,message := range m.Messages {
		fmt.Println(message.ProductID,"   ",message.UserID)
	}
}