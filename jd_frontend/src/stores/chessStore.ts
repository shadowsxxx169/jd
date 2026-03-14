import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// 后端棋盘类型：10行 x 9列字符串数组
export type BackendBoard = string[][]

// 坐标转换工具：后端 [row,col] ↔ xiangqiboard 方格名（如 'e9'）
const COL_LETTERS = 'abcdefghi'

export function backendToSquare(row: number, col: number): string {
  return COL_LETTERS[col] + (9 - row)
}

export function squareToBackend(square: string): [number, number] {
  const col = square.charCodeAt(0) - 97
  const row = 9 - parseInt(square[1])
  return [row, col]
}

// 棋子代码转换：后端 H/E ↔ xiangqiboard N/B
const TO_BOARD_PIECE: Record<string, string> = { H: 'N', E: 'B' }
const TO_BACKEND_PIECE: Record<string, string> = { N: 'H', B: 'E' }

export function backendPieceToBoard(piece: string): string {
  if (!piece) return piece
  return piece[0] + (TO_BOARD_PIECE[piece[1]] ?? piece[1])
}

export function boardPieceToBackend(piece: string): string {
  if (!piece) return piece
  return piece[0] + (TO_BACKEND_PIECE[piece[1]] ?? piece[1])
}

// 后端棋盘 → xiangqiboard position 对象
export function backendBoardToPosition(board: BackendBoard): Record<string, string> {
  const pos: Record<string, string> = {}
  for (let r = 0; r < 10; r++) {
    for (let c = 0; c < 9; c++) {
      const piece = board[r][c]
      if (piece) {
        pos[backendToSquare(r, c)] = backendPieceToBoard(piece)
      }
    }
  }
  return pos
}

export type GameStatus = 'lobby' | 'waiting' | 'playing' | 'ended'

export const useChessStore = defineStore('chess', () => {
  // 游戏状态
  const status = ref<GameStatus>('lobby')
  const roomId = ref('')
  const myColor = ref<'red' | 'black' | ''>('')
  const turn = ref<'red' | 'black'>('red')
  const winner = ref('')
  const message = ref('')
  const board = ref<BackendBoard>([])

  // 走棋记录
  const moveHistory = ref<Array<{ from: string; to: string; piece: string }>>([])

  // WebSocket 实例
  let ws: WebSocket | null = null

  // 获取 WS 地址
  function getWsUrl(): string {
    const wsBase = import.meta.env.VITE_WS_URL
    if (wsBase) return wsBase
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${proto}//${location.host}/ws/chess`
  }

  // 连接 WebSocket
  function connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      ws = new WebSocket(getWsUrl())
      ws.onopen = () => resolve()
      ws.onerror = () => reject(new Error('WebSocket 连接失败'))
      ws.onmessage = (evt) => handleMessage(JSON.parse(evt.data))
      ws.onclose = () => {
        if (status.value === 'playing') {
          message.value = '连接已断开'
        }
      }
    })
  }

  // 处理服务端消息
  function handleMessage(msg: any) {
    switch (msg.type) {
      case 'room_created':
        roomId.value = msg.room_id
        status.value = 'waiting'
        message.value = `房间已创建，等待对手加入…  房间号：${msg.room_id}`
        break

      case 'game_start':
        board.value = msg.board
        turn.value = msg.turn
        status.value = 'playing'
        message.value = '游戏开始！红方先走'
        break

      case 'your_color':
        myColor.value = msg.color
        break

      case 'move':
        board.value = msg.board
        turn.value = msg.turn
        if (msg.from && msg.to) {
          moveHistory.value.push({
            from: backendToSquare(msg.from.row, msg.from.col),
            to: backendToSquare(msg.to.row, msg.to.col),
            piece: msg.piece ?? '',
          })
        }
        message.value = turn.value === 'red' ? '红方走棋' : '黑方走棋'
        break

      case 'error':
        message.value = msg.message ?? '操作失败'
        break

      case 'game_over':
        winner.value = msg.winner
        status.value = 'ended'
        message.value =
          msg.winner === 'red'
            ? '红方胜利！'
            : msg.winner === 'black'
              ? '黑方胜利！'
              : '平局'
        break

      case 'opponent_disconnected':
        message.value = '对手已断线'
        status.value = 'ended'
        break
    }
  }

  // 创建房间
  async function createRoom() {
    await connect()
    ws!.send(JSON.stringify({ type: 'create_room' }))
  }

  // 加入房间
  async function joinRoom(id: string) {
    await connect()
    ws!.send(JSON.stringify({ type: 'join_room', room_id: id.toUpperCase() }))
  }

  // 发送落子
  function sendMove(source: string, target: string) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return
    const [fromRow, fromCol] = squareToBackend(source)
    const [toRow, toCol] = squareToBackend(target)
    ws.send(
      JSON.stringify({
        type: 'move',
        from: { row: fromRow, col: fromCol },
        to: { row: toRow, col: toCol },
      }),
    )
  }

  // 认输
  function resign() {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'resign' }))
    }
  }

  // 重置回大厅
  function reset() {
    ws?.close()
    ws = null
    status.value = 'lobby'
    roomId.value = ''
    myColor.value = ''
    turn.value = 'red'
    winner.value = ''
    message.value = ''
    board.value = []
    moveHistory.value = []
  }

  // 是否轮到我走棋
  const isMyTurn = computed(() => turn.value === myColor.value)

  return {
    status,
    roomId,
    myColor,
    turn,
    winner,
    message,
    board,
    moveHistory,
    isMyTurn,
    createRoom,
    joinRoom,
    sendMove,
    resign,
    reset,
    backendBoardToPosition,
  }
})
