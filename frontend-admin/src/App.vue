<script setup>
// App entry - router-view will display current route component
</script>

<template>
  <router-view />
</template>

<style>
/* 全局样式重置 */
html, body {
  margin: 0;
  padding: 0;
  height: 100%;
  width: 100%;
  overflow-y: auto;
}

#app {
  height: 100%;
  width: 100%;
  max-width: 100%;
  margin: 0;
  padding: 0;
  text-align: left;
  overflow-y: auto;
}

/* 确保内容可滚动 */
.el-container, .el-main {
  height: auto;
  min-height: 100%;
  overflow-y: auto;
}

.el-card {
  height: auto;
  overflow: visible;
}

.el-card__body {
  height: auto;
  overflow: visible;
}
</style>
