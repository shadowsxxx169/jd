<script lang="ts" setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useChessStore } from '@/stores/chessStore'

const router = useRouter()
const store = useChessStore()

const joinInput = ref('')
const loading = ref(false)
const error = ref('')

// 创建房间
async function handleCreate() {
  loading.value = true
  error.value = ''
  try {
    await store.createRoom()
    router.push('/chess/game')
  } catch (e: any) {
    error.value = e.message ?? '创建失败'
  } finally {
    loading.value = false
  }
}

// 加入房间
async function handleJoin() {
  if (!joinInput.value.trim()) {
    error.value = '请输入房间号'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await store.joinRoom(joinInput.value.trim())
    router.push('/chess/game')
  } catch (e: any) {
    error.value = e.message ?? '加入失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-[#1a0a00] flex flex-col items-center justify-center px-4">
    <!-- 标题 -->
    <div class="mb-10 text-center">
      <h1 class="text-5xl font-bold text-[#c8960c] tracking-widest mb-2">中国象棋</h1>
      <p class="text-[#8b6914] text-lg">在线双人对弈</p>
    </div>

    <!-- 卡片 -->
    <div class="w-full max-w-sm bg-[#2d1500] border border-[#5a3000] rounded-2xl p-8 shadow-2xl">

      <!-- 创建房间 -->
      <button
        class="w-full py-4 mb-4 rounded-xl text-xl font-bold text-white bg-[#c41e3a] hover:bg-[#a01830] transition disabled:opacity-50"
        :disabled="loading"
        @click="handleCreate"
      >
        {{ loading ? '连接中…' : '创建房间' }}
      </button>

      <div class="flex items-center gap-3 my-5">
        <div class="flex-1 h-px bg-[#5a3000]" />
        <span class="text-[#8b6914] text-sm">或加入已有房间</span>
        <div class="flex-1 h-px bg-[#5a3000]" />
      </div>

      <!-- 加入房间 -->
      <input
        v-model="joinInput"
        class="w-full px-4 py-3 mb-3 rounded-xl bg-[#1a0a00] border border-[#5a3000] text-[#f0c060] text-xl text-center tracking-[0.4em] uppercase placeholder:text-[#5a3000] focus:outline-none focus:border-[#c8960c]"
        placeholder="房间号"
        maxlength="4"
        @keyup.enter="handleJoin"
      />
      <button
        class="w-full py-3 rounded-xl text-lg font-bold text-[#1a0a00] bg-[#c8960c] hover:bg-[#f0c060] transition disabled:opacity-50"
        :disabled="loading || !joinInput.trim()"
        @click="handleJoin"
      >
        加入对局
      </button>

      <!-- 错误提示 -->
      <p v-if="error" class="mt-4 text-center text-red-400 text-sm">{{ error }}</p>
    </div>

    <!-- 规则说明 -->
    <div class="mt-8 text-[#5a3000] text-sm text-center max-w-xs leading-6">
      创建房间后将获得 4 位房间号，<br />
      将房间号分享给对手即可开始对局
    </div>
  </div>
</template>
