<script setup lang="ts">
import { ref } from 'vue'
import type { MetaObject } from '@/types/file'

defineProps<{ files: MetaObject[] }>()
const emit = defineEmits<{
  download: [file: MetaObject]
  openFolder: [folder: MetaObject]
  move: [itemId: number, targetFolderId: number]
}>()

const dragOverFolderId = ref<number | null>(null)
let draggingId: number | null = null

function onDragStart(file: MetaObject) {
  draggingId = file.id
}

function onDragOver(folder: MetaObject) {
  if (draggingId !== null && draggingId !== folder.id) {
    dragOverFolderId.value = folder.id
  }
}

function onDragLeave() {
  dragOverFolderId.value = null
}

function onDrop(folder: MetaObject) {
  if (draggingId !== null && draggingId !== folder.id) {
    emit('move', draggingId, folder.id)
  }
  draggingId = null
  dragOverFolderId.value = null
}

function onDblClick(file: MetaObject) {
  if (file.is_file) {
    emit('download', file)
  } else {
    emit('openFolder', file)
  }
}
</script>

<template>
  <div class="file-grid">
    <div
      v-for="file in files"
      :key="file.id"
      class="file-card"
      :class="{
        deleted: file.deleted_at !== null,
        'drag-over': !file.is_file && dragOverFolderId === file.id,
      }"
      :draggable="true"
      @dragstart="onDragStart(file)"
      @dragover.prevent="!file.is_file ? onDragOver(file) : undefined"
      @dragleave="onDragLeave"
      @drop.prevent="!file.is_file ? onDrop(file) : undefined"
      @dblclick="onDblClick(file)"
    >
      <div v-if="file.is_file" class="file-icon-rect">
        <div class="fold" />
      </div>
      <div v-else class="folder-icon-rect" />

      <div class="file-name">{{ file.name }}</div>
      <div class="file-type">{{ file.file_type }}</div>
      <div class="file-version">v{{ file.version }}</div>
      <div v-if="file.deleted_at !== null" class="deleted-badge">Deleted</div>
    </div>
    <div v-if="files.length === 0" class="empty-state">
      <p>No files yet</p>
    </div>
  </div>
</template>

<style scoped>
.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 20px;
  padding: 20px;
}

.file-card {
  background: #fff;
  border: 2px solid #e0e0e0;
  border-radius: 12px;
  padding: 20px 16px 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  min-height: 180px;
  user-select: none;
}

.file-card:hover {
  border-color: #4CAF50;
  transform: translateY(-3px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.file-card.deleted { opacity: 0.45; border-color: #f44336; }

.file-card.drag-over {
  border-color: #1976d2;
  background: #e3f2fd;
  transform: translateY(-3px);
  box-shadow: 0 4px 16px rgba(25, 118, 210, 0.25);
}

/* File icon: rectangle with folded top-right corner */
.file-icon-rect {
  width: 44px;
  height: 56px;
  background: #e3f2fd;
  border: 2px solid #90caf9;
  border-radius: 3px 0 3px 3px;
  position: relative;
  flex-shrink: 0;
}

.fold {
  position: absolute;
  top: -2px;
  right: -2px;
  width: 14px;
  height: 14px;
  background: white;
  border-left: 2px solid #90caf9;
  border-bottom: 2px solid #90caf9;
  border-radius: 0 0 0 3px;
}

/* Folder icon */
.folder-icon-rect {
  width: 52px;
  height: 42px;
  background: #fff9c4;
  border: 2px solid #f9a825;
  border-radius: 0 4px 4px 4px;
  position: relative;
  flex-shrink: 0;
}

.folder-icon-rect::before {
  content: '';
  position: absolute;
  top: -10px;
  left: -2px;
  width: 22px;
  height: 10px;
  background: #fff9c4;
  border: 2px solid #f9a825;
  border-bottom: none;
  border-radius: 4px 4px 0 0;
}

.file-name {
  font-weight: 600;
  font-size: 13px;
  text-align: center;
  word-break: break-all;
  color: #333;
  width: 100%;
}

.file-type { font-size: 11px; color: #888; text-align: center; }

.file-version {
  font-size: 10px;
  color: #aaa;
  background: #f5f5f5;
  padding: 2px 8px;
  border-radius: 10px;
}

.deleted-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  background: #f44336;
  color: white;
  font-size: 10px;
  padding: 3px 7px;
  border-radius: 4px;
  font-weight: 600;
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
  color: #bbb;
  font-size: 18px;
}
</style>
