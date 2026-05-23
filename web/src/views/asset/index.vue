<template>
  <div style="padding: 16px">
    <a-card :bordered="false">
      <template #title>资产管理</template>
      <a-table :columns="columns" :data="assets" />
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const assets = ref([])
const columns = [
  { title: 'ID', dataIndex: 'id', width: 80 },
  { title: '名称', dataIndex: 'name' },
  { title: '类型', dataIndex: 'type', width: 100 },
  { title: 'IP', dataIndex: 'ip', width: 140 },
  { title: '状态', dataIndex: 'status', width: 80 },
  { title: '位置', dataIndex: 'location', width: 200 },
  { title: '操作', slotName: 'action', width: 120 },
]
</script>
