package service

import (
	"cchat/pkg/protocol"
	"log"
	"strconv"
	"sync"

	"github.com/gogo/protobuf/proto"
)

type Server struct {
	Clients   map[string]*Client
	mutex     *sync.Mutex
	Broadcast chan []byte
	Register  chan *Client
	Ungister  chan *Client
}

func NewServer() *Server {
	return &Server{
		Clients:   make(map[string]*Client),
		Broadcast: make(chan []byte), // 无缓冲
		Register:  make(chan *Client),
		Ungister:  make(chan *Client),
	}
}

var ServerInstance = NewServer()

func (s *Server) Start() {
	for {
		select {
		case conn := <-s.Register:
			log.Println("注册新连接:", conn.UUID)
			s.Clients[conn.UUID] = conn
		case conn := <-s.Ungister:
			log.Println("注销连接:", conn.UUID)
			delete(s.Clients, conn.UUID)
		case message := <-s.Broadcast:
			log.Println("收发消息:")
			msg := protocol.Message{}
			proto.Unmarshal(message, &msg)
			// 表示有意向的消息
			if msg.To != "" {
				log.Println("单聊消息")
				switch msg.ContentType {
				case 1: //Text消息
					if msg.MessageType == 1 {
						client, ok := s.Clients[msg.To]
						if ok {
							msgByte, err := proto.Marshal(&msg)
							if err != nil {
								return
							}
							log.Println("单聊消息给：", msg.To)
							client.Send <- msgByte
						}
					} else {
						// 群聊消息
						msgByte, err := proto.Marshal(&msg)
						if err != nil {
							return
						}
						log.Println("单聊消息给：", msg.To)
						s.SendGroupMessage(msg.To, msgByte)
					}
				case 2: // 文件消息
					if msg.MessageType == 1 {
						client, ok := s.Clients[msg.To]
						if ok {
							msgByte, err := proto.Marshal(&msg)
							if err != nil {
								return
							}
							log.Println("发送文件消息给：", msg.To)
							client.Send <- msgByte
						}
					}
				case 3: // 图片消息
					if msg.MessageType == 1 {
						client, ok := s.Clients[msg.To]
						if ok {
							msgByte, err := proto.Marshal(&msg)
							if err != nil {
								return
							}
							log.Println("发送图片消息给：", msg.To)
							client.Send <- msgByte
						}
					}
				case 4: // 语音消息
					if msg.MessageType == 1 {
						client, ok := s.Clients[msg.To]
						if ok {
							msgByte, err := proto.Marshal(&msg)
							if err != nil {
								return
							}
							log.Println("发送语音消息给：", msg.To)
							client.Send <- msgByte
						}
					}
				case 8: // 加好友消息
					if msg.MessageType == 1 {
						client, ok := s.Clients[msg.To]
						if ok {
							msgByte, err := proto.Marshal(&msg)
							if err != nil {
								return
							}
							log.Println("发送好友消息给：", msg.To)
							client.Send <- msgByte
						}
					}
				}
			} else {
				log.Println("群聊消息")
				for _, client := range s.Clients {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(s.Clients, client.UUID)
					}
				}
			}
		}
	}
}

func (s *Server) GetClient(uuid string) *Client {
	return s.Clients[uuid]
}

func (s *Server) SendGroupMessage(to string, msg []byte) {
	// 获取该群聊下的所有群成员
	groupService := &GroupService{}
	// string to int
	groupId, err := strconv.Atoi(to)
	if err != nil {
		log.Println("群聊ID转换失败")
		return
	}
	groupMembers, err := groupService.GetGroupMember(groupId)
	if err != nil {
		log.Println("获取群聊下的所有群成员失败")
		return
	}
	for _, clientID := range groupMembers {
		client, ok := s.Clients[clientID]
		if ok {
			client.Send <- msg
		}
	}
}
