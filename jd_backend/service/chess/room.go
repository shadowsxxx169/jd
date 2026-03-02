package chess

import (
	"math/rand"
	"sync"
	"time"
)

// RoomStatus 房间状态
type RoomStatus string

const (
	RoomWaiting RoomStatus = "waiting" // 等待第二名玩家
	RoomPlaying RoomStatus = "playing" // 游戏进行中
	RoomEnded   RoomStatus = "ended"   // 游戏结束
)

// MoveRecord 走棋记录
type MoveRecord struct {
	From  Pos    `json:"from"`
	To    Pos    `json:"to"`
	Piece string `json:"piece"`
}

// Room 象棋对局房间
type Room struct {
	ID        string        `json:"id"`
	Board     Board         `json:"board"`
	Turn      string        `json:"turn"`   // "red" 或 "black"
	Status    RoomStatus    `json:"status"` // 房间状态
	Winner    string        `json:"winner"` // 胜者，"red"/"black"/"draw"
	MoveList  []MoveRecord  `json:"move_list"`
	RedConn   *ClientConn   `json:"-"` // 红方连接
	BlackConn *ClientConn   `json:"-"` // 黑方连接
	mu        sync.Mutex
}

// ClientConn 客户端连接抽象（避免循环依赖，用接口）
type ClientConn interface {
	SendJSON(v any) error
	Close() error
}

// Manager 全局房间管理器
type Manager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewManager 创建房间管理器
func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

// 生成4位随机房间ID
func genRoomID() string {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 4)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

// CreateRoom 创建新房间，返回房间ID
func (m *Manager) CreateRoom() *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 生成不重复的房间ID
	id := genRoomID()
	for m.rooms[id] != nil {
		id = genRoomID()
	}

	room := &Room{
		ID:       id,
		Board:    InitialBoard(),
		Turn:     "red",
		Status:   RoomWaiting,
		MoveList: []MoveRecord{},
	}
	m.rooms[id] = room
	return room
}

// GetRoom 根据ID获取房间
func (m *Manager) GetRoom(id string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	room, ok := m.rooms[id]
	return room, ok
}

// RemoveRoom 删除房间
func (m *Manager) RemoveRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

// JoinRoom 加入房间，返回分配的颜色（"red"/"black"）和是否成功
func (m *Manager) JoinRoom(id string, conn ClientConn) (string, *Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[id]
	if !ok {
		return "", nil, false
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Status != RoomWaiting {
		return "", nil, false
	}

	// 红方已有人，加入为黑方
	if room.RedConn != nil && room.BlackConn == nil {
		room.BlackConn = &conn
		room.Status = RoomPlaying
		return "black", room, true
	}
	return "", nil, false
}

// SetRedConn 设置红方连接（建房时调用）
func (r *Room) SetRedConn(conn ClientConn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RedConn = &conn
}

// Broadcast 向房间内所有玩家广播消息
func (r *Room) Broadcast(msg any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.RedConn != nil {
		(*r.RedConn).SendJSON(msg)
	}
	if r.BlackConn != nil {
		(*r.BlackConn).SendJSON(msg)
	}
}

// SendTo 向指定方发送消息
func (r *Room) SendTo(color string, msg any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if color == "red" && r.RedConn != nil {
		(*r.RedConn).SendJSON(msg)
	}
	if color == "black" && r.BlackConn != nil {
		(*r.BlackConn).SendJSON(msg)
	}
}

// DoMove 执行落子，返回是否成功及游戏是否结束
func (r *Room) DoMove(color string, from, to Pos) (ok bool, gameOver bool, winner string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 不是该方的回合
	if r.Turn != color {
		return false, false, ""
	}
	// 游戏未在进行中
	if r.Status != RoomPlaying {
		return false, false, ""
	}
	// 校验合法性
	if !IsValidMove(r.Board, from, to) {
		return false, false, ""
	}

	// 记录并执行
	piece := r.Board[from.Row][from.Col]
	r.Board = ApplyMove(r.Board, from, to)
	r.MoveList = append(r.MoveList, MoveRecord{from, to, piece})

	// 切换回合
	next := "black"
	if color == "black" {
		next = "red"
	}
	r.Turn = next

	// 检测将死/困毙
	if !HasAnyMoves(r.Board, next) {
		r.Status = RoomEnded
		r.Winner = color
		return true, true, color
	}

	return true, false, ""
}

// Resign 认输
func (r *Room) Resign(color string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Status = RoomEnded
	if color == "red" {
		r.Winner = "black"
	} else {
		r.Winner = "red"
	}
	return r.Winner
}
