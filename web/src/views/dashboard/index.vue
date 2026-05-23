<template>
  <div class="dashboard">
    <a-row :gutter="16">
      <a-col :span="6">
        <a-card :bordered="false">
          <a-statistic title="待处理工单" :value="stats.pending" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false">
          <a-statistic title="处理中工单" :value="stats.processing" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false">
          <a-statistic title="本月完单" :value="stats.completed" />
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card :bordered="false">
          <a-statistic title="SLA达成率" :value="stats.sla_rate" suffix="%" />
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" style="margin-top: 16px">
      <a-col :span="16">
        <a-card title="工单趋势" :bordered="false">
          <div ref="trendChartRef" style="height: 300px"></div>
        </a-card>
      </a-col>
      <a-col :span="8">
        <a-card title="工单分布" :bordered="false">
          <div ref="pieChartRef" style="height: 300px"></div>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'

const trendChartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()

const stats = reactive({
  pending: 0,
  processing: 0,
  completed: 0,
  sla_rate: 0,
})

onMounted(() => {
  // TODO: fetch dashboard data and render charts with ECharts
})
</script>

<style scoped>
.dashboard {
  padding: 16px;
}
</style>
