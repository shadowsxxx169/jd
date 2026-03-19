<script lang="ts" setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useChessStore, backendBoardToPosition } from '@/stores/chessStore'

// xiangqiboardjs 通过 index.html 中的 <script> 标签挂载到 window
declare const Xiangqiboard: any

const router = useRouter()
const store = useChessStore()

const boardEl = ref<HTMLElement | null>(null)
let board: any = null // xiangqiboard 实例

// 初始化棋盘（在 game_start 后调用）
function initBoard(position: Record<string, string>) {
  if (!boardEl.value) return
  if (board) board.destroy()

  board = Xiangqiboard(boardEl.value, {
    position,
    orientation: store.myColor === 'black' ? 'black' : 'red',
    draggable: true,
    // 棋子图片路径：public/xiangqi/img/... 由 Vite 原样暴露
    pieceTheme: '/xiangqi/img/xiangqipieces/wikimedia/{piece}.svg',
    dropOffBoard: 'snapback',

    // 拖动开始：不是我方棋子或不是我的回合 → 阻止
    onDragStart(_source: string, piece: string) {
      if (store.status !== 'playing') return false
      if (!store.isMyTurn) return false
      // 只能拖动自己颜色的棋子
      const myPrefix = store.myColor === 'red' ? 'r' : 'b'
      if (!piece.startsWith(myPrefix)) return false
    },

    // 落子：发送给服务端，由服务端决定是否合法
    onDrop(source: string, target: string) {
      if (source === target) return 'snapback'
      // 乐观接受，等服务端响应后同步棋盘
      store.sendMove(source, target)
    },

    // 拖动结束后立即刷新为当前 store 状态（防止乐观更新错位）
    onSnapEnd() {
      if (board && store.board.length) {
        board.position(backendBoardToPosition(store.board), false)
      }
    },
  })
}

// 监听棋盘变化，同步到 xiangqiboardjs
watch(
  () => store.board,
  (newBoard) => {
    if (!newBoard.length) return
    if (!board) {
      // 首次收到棋盘数据时初始化
      initBoard(backendBoardToPosition(newBoard))
    } else {
      board.position(backendBoardToPosition(newBoard), true)
    }
  },
  { deep: true },
)

// 游戏结束后不需要额外操作，状态已在 store 中

onMounted(() => {
  // 如果刷新页面直接进入 /chess/game（没有房间），跳回大厅
  if (store.status === 'lobby') {
    router.replace('/chess')
  }
})

onUnmounted(() => {
  board?.destroy()
  board = null
})

function handleResign() {
  if (confirm('确定认输？')) store.resign()
}

function handleBack() {
  store.reset()
  router.push('/chess')
}

// 棋子名称映射（用于走棋记录显示）
const PIECE_NAMES: Record<string, string> = {
  rK: '帅', rA: '仕', rB: '相', rN: '马', rR: '车', rC: '炮', rP: '兵',
  bK: '将', bA: '士', bB: '象', bN: '馬', bR: '車', bC: '砲', bP: '卒',
}
</script>

<template>
  <div class="min-h-screen bg-[#1a0a00] flex flex-col items-center py-6 px-4">

    <!-- 顶部状态栏 -->
    <div class="w-full max-w-2xl flex items-center justify-between mb-4">
      <button
        class="text-[#8b6914] hover:text-[#c8960c] text-sm transition"
        @click="handleBack"
      >
        ← 返回大厅
      </button>

      <div class="text-center">
        <span
          v-if="store.roomId"
          class="text-[#8b6914] text-sm"
        >
          房间号：<span class="text-[#c8960c] font-bold tracking-widest">{{ store.roomId }}</span>
        </span>
      </div>

      <!-- 认输按钮 -->
      <button
        v-if="store.status === 'playing'"
        class="text-sm text-red-500 hover:text-red-400 transition border border-red-800 px-3 py-1 rounded-lg"
        @click="handleResign"
      >
        认输
      </button>
      <div v-else class="w-16" />
    </div>

    <!-- 游戏消息提示 -->
    <div
      class="mb-4 px-6 py-2 rounded-full text-sm font-medium transition-all"
      :class="{
        'bg-[#c41e3a] text-white': store.status === 'ended',
        'bg-[#2d1500] text-[#c8960c]': store.status !== 'ended',
      }"
    >
      {{ store.message || '等待连接…' }}
    </div>

    <!-- 主区域：棋盘 + 走棋记录 -->
    <div class="flex gap-6 items-start">

      <!-- 棋盘容器 -->
      <div class="relative">
        <!-- 黑方标识 -->
        <div
          class="flex items-center gap-2 mb-2 justify-center"
          :class="store.myColor === 'black' ? 'text-[#c8960c]' : 'text-[#5a3000]'"
        >
          <div class="w-3 h-3 rounded-full bg-current" />
          <span class="text-sm font-bold">{{ store.myColor === 'black' ? '我方（黑）' : '对手（黑）' }}</span>
        </div>

        <!-- 棋盘 -->
        <div
          ref="boardEl"
          style="width: 400px"
          class="shadow-2xl rounded-md overflow-hidden"
        />

        <!-- 红方标识 -->
        <div
          class="flex items-center gap-2 mt-2 justify-center"
          :class="store.myColor === 'red' ? 'text-[#c8960c]' : 'text-[#5a3000]'"
        >
          <div class="w-3 h-3 rounded-full bg-[#c41e3a]" />
          <span class="text-sm font-bold">{{ store.myColor === 'red' ? '我方（红）' : '对手（红）' }}</span>
        </div>

        <!-- 等待对手覆盖层 -->
        <div
          v-if="store.status === 'waiting'"
          class="absolute inset-0 bg-black/70 flex flex-col items-center justify-center rounded-md"
        >
          <div class="text-[#c8960c] text-4xl mb-3 animate-pulse">⏳</div>
          <p class="text-white text-lg font-bold">等待对手加入</p>
          <p class="text-[#8b6914] text-sm mt-2">房间号：<span class="text-[#c8960c] tracking-widest font-bold">{{ store.roomId }}</span></p>
        </div>
      </div>

      <!-- 走棋记录面板 -->
      <div class="w-36 bg-[#2d1500] border border-[#5a3000] rounded-xl p-4 h-[480px] flex flex-col">
        <h3 class="text-[#c8960c] text-sm font-bold mb-3 text-center">走棋记录</h3>
        <div class="flex-1 overflow-y-auto space-y-1 text-xs">
          <div
            v-for="(move, idx) in store.moveHistory"
            :key="idx"
            class="flex items-center gap-1 text-[#8b6914]"
          >
            <span class="text-[#5a3000] w-5 text-right">{{ idx + 1 }}.</span>
            <span
              class="font-bold"
              :class="move.piece.startsWith('r') ? 'text-[#c41e3a]' : 'text-gray-300'"
            >
              {{ PIECE_NAMES[move.piece] ?? move.piece }}
            </span>
            <span>{{ move.from }}→{{ move.to }}</span>
          </div>
          <p v-if="!store.moveHistory.length" class="text-[#5a3000] text-center mt-4">暂无记录</p>
        </div>
      </div>
    </div>

    <!-- 游戏结束弹层 -->
    <div
      v-if="store.status === 'ended'"
      class="fixed inset-0 bg-black/60 flex items-center justify-center z-50"
      @click.self="handleBack"
    >
      <div class="bg-[#2d1500] border-2 border-[#c8960c] rounded-2xl p-10 text-center shadow-2xl">
        <div class="text-6xl mb-4">{{ store.winner === store.myColor ? '🏆' : '😔' }}</div>
        <h2 class="text-3xl font-bold text-[#c8960c] mb-2">{{ store.message }}</h2>
        <button
          class="mt-6 px-8 py-3 rounded-xl bg-[#c8960c] text-[#1a0a00] font-bold hover:bg-[#f0c060] transition"
          @click="handleBack"
        >
          返回大厅
        </button>
      </div>
    </div>

  </div>
</template>
