<template>
  <div style="padding: 16px">
    <a-card :bordered="false">
      <template #title>知识库</template>
      <a-table :columns="columns" :data="articles" />
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const articles = ref([])
const columns = [
  { title: 'ID', dataIndex: 'id', width: 80 },
  { title: '标题', dataIndex: 'title' },
  { title: '标签', dataIndex: 'tags', width: 200 },
  { title: '浏览量', dataIndex: 'view_count', width: 80 },
  { title: '创建时间', dataIndex: 'created_at', width: 180 },
  { title: '操作', slotName: 'action', width: 120 },
]
</script>
