package controller

import (
	"jd/service/chess"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// upgrader 将 HTTP 连接升级为 WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应校验 Origin
	},
}

// wsMsg 客户端发来的消息结构
type wsMsg struct {
	Type   string    `json:"type"`
	RoomID string    `json:"room_id,omitempty"`
	From   chess.Pos `json:"from,omitempty"`
	To     chess.Pos `json:"to,omitempty"`
}

// outMsg 服务端发出的消息结构
type outMsg struct {
	Type    string      `json:"type"`
	RoomID  string      `json:"room_id,omitempty"`
	Color   string      `json:"color,omitempty"`
	Board   chess.Board `json:"board,omitempty"`
	Turn    string      `json:"turn,omitempty"`
	From    *chess.Pos  `json:"from,omitempty"`
	To      *chess.Pos  `json:"to,omitempty"`
	Piece   string      `json:"piece,omitempty"`
	Winner  string      `json:"winner,omitempty"`
	Message string      `json:"message,omitempty"`
}

// wsClient 封装 WebSocket 连接，保证并发安全
type wsClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *wsClient) SendJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(v)
}

func (c *wsClient) Close() error {
	return c.conn.Close()
}

// ChessController 象棋 WebSocket 控制器
type ChessController struct {
	manager *chess.Manager
}

// NewChessController 创建控制器，注入全局房间管理器
func NewChessController(m *chess.Manager) *ChessController {
	return &ChessController{manager: m}
}

// HandleWS 处理 WebSocket 连接 GET /ws/chess
func (cc *ChessController) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}
	defer conn.Close()

	client := &wsClient{conn: conn}

	var (
		myRoom  *chess.Room
		myColor string
	)

	// 消息循环：持续读取客户端消息
	for {
		var msg wsMsg
		if err := conn.ReadJSON(&msg); err != nil {
			// 连接断开，通知对手
			if myRoom != nil {
				myRoom.Broadcast(outMsg{
					Type:    "opponent_disconnected",
					Message: "对手已断开连接",
				})
			}
			break
		}

		switch msg.Type {

		case "create_room":
			// 创建房间，成为红方
			myRoom = cc.manager.CreateRoom()
			myRoom.SetRedConn(client)
			myColor = "red"
			client.SendJSON(outMsg{
				Type:   "room_created",
				RoomID: myRoom.ID,
				Color:  "red",
			})

		case "join_room":
			// 加入已有房间，成为黑方
			color, r, ok := cc.manager.JoinRoom(msg.RoomID, client)
			if !ok {
				client.SendJSON(outMsg{
					Type:    "error",
					Message: "房间不存在或已满员",
				})
				continue
			}
			myColor = color
			myRoom = r
			// 通知双方游戏开始
			myRoom.Broadcast(outMsg{
				Type:  "game_start",
				Board: myRoom.Board,
				Turn:  "red",
			})
			// 分别告知各自颜色
			myRoom.SendTo("red", outMsg{Type: "your_color", Color: "red"})
			myRoom.SendTo("black", outMsg{Type: "your_color", Color: "black"})

		case "move":
			// 落子请求
			if myRoom == nil {
				client.SendJSON(outMsg{Type: "error", Message: "尚未进入房间"})
				continue
			}
			fromPos := msg.From
			toPos := msg.To
			ok, gameOver, winner := myRoom.DoMove(myColor, fromPos, toPos)
			if !ok {
				client.SendJSON(outMsg{Type: "error", Message: "非法走法"})
				continue
			}
			// 广播落子结果
			myRoom.Broadcast(outMsg{
				Type:  "move",
				Board: myRoom.Board,
				Turn:  myRoom.Turn,
				From:  &fromPos,
				To:    &toPos,
				Piece: myRoom.Board[toPos.Row][toPos.Col],
			})
			// 游戏结束则广播结果
			if gameOver {
				myRoom.Broadcast(outMsg{
					Type:   "game_over",
					Winner: winner,
				})
			}

		case "resign":
			// 认输
			if myRoom == nil {
				continue
			}
			winner := myRoom.Resign(myColor)
			myRoom.Broadcast(outMsg{
				Type:    "game_over",
				Winner:  winner,
				Message: myColor + " 认输",
			})

		default:
			client.SendJSON(outMsg{Type: "error", Message: "未知消息类型: " + msg.Type})
		}
	}
}
