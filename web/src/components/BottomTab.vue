<template>
  <nav class="bottom-tab">
    <router-link
      to="/"
      class="tab-item"
      active-class="active"
      :exact-active-class="'active'"
    >
      <svg class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 10.5 12 4l9 6.5"/><path d="M5 9.5V21h14V9.5"/></svg>
      <span class="tab-label">首页</span>
    </router-link>

    <router-link to="/ranking" class="tab-item" active-class="active">
      <svg class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M4 20V10m6 10V4m6 16v-7M20 14V6"/></svg>
      <span class="tab-label">排行榜</span>
    </router-link>

    <router-link to="/me" class="tab-item" active-class="active">
      <svg class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="12" cy="8" r="4"/><path d="M4 21c0-3.3 3.6-6 8-6s8 2.7 8 6"/></svg>
      <span class="tab-label">我的</span>
    </router-link>
  </nav>
</template>

<script setup>
</script>

<style scoped>
.bottom-tab {
  position: fixed;
  left: 0; right: 0; bottom: 0;
  height: var(--tabbar-height, 60px);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2xl);
  background: rgba(16,18,34,0.55);
  backdrop-filter: blur(60px) saturate(160%);
  -webkit-backdrop-filter: blur(60px) saturate(160%);
  border-top: 0.5px solid var(--color-border);
  z-index: 150;
  padding-bottom: env(safe-area-inset-bottom, 0);
}

.tab-item {
  width: 76px;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  color: var(--color-text-quaternary);
  transition: all var(--transition-fast);
  position: relative;
}

.tab-icon {
  width: 22px;
  height: 22px;
  transition: transform var(--transition-fast);
}

.tab-label {
  font-size: 10px;
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.02em;
}

.tab-item.active {
  color: var(--glass-blue);
}
.tab-item.active .tab-icon {
  transform: translateY(-1px) scale(1.05);
}

/* active 指示圆点 */
.tab-item.active::after {
  content: '';
  position: absolute;
  top: 4px;
  left: 50%;
  transform: translateX(-50%);
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--glass-blue);
  box-shadow: 0 0 8px var(--glass-blue);
}
</style>