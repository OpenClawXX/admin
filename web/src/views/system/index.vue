<template>
  <div style="padding: 16px">
    <a-card :bordered="false">
      <template #title>系统设置</template>
      <a-tabs>
        <a-tab-pane key="sla" title="SLA 配置">
          <p>SLA 规则配置（开发中）</p>
        </a-tab-pane>
        <a-tab-pane key="score" title="评分权重">
          <p>评分维度权重配置（开发中）</p>
        </a-tab-pane>
        <a-tab-pane key="notify" title="通知配置">
          <p>通知推送配置（开发中）</p>
        </a-tab-pane>
      </a-tabs>
    </a-card>
  </div>
</template>

<script setup lang="ts">
</script>
