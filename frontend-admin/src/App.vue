<script setup>
// App entry - router-view will display current route component
</script>

<template>
  <router-view />
</template>

<style scoped>
/* Add global styles if necessary */
.logo {
  height: 6em;
  padding: 1.5em;
  will-change: filter;
  transition: filter 300ms;
}
.logo:hover {
  filter: drop-shadow(0 0 2em #646cffaa);
}
.logo.vue:hover {
  filter: drop-shadow(0 0 2em #42b883aa);
}
</style>
