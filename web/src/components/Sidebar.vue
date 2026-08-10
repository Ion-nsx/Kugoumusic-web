<template>
  <aside class="sidebar">
    <!-- 品牌 Logo -->
    <div class="logo">
      <div class="logo-badge">
        <img src="/favicon.svg" alt="酷狗音乐概念" />
      </div>
      <div class="logo-name">酷狗音乐<span>概念</span></div>
    </div>

    <!-- 推荐 -->
    <div class="nav-group">发现音乐</div>
    <router-link to="/" class="nav-item" active-class="active" exact-active-class="active">
      <span class="nav-ic"><svg viewBox="0 0 24 24"><path d="M3 10.5 12 3l9 7.5"/><path d="M5 9.5V21h14V9.5"/></svg></span>
      <span>发现音乐</span>
    </router-link>
    <router-link to="/daily" class="nav-item" active-class="active">
      <span class="nav-ic"><svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/><circle cx="12" cy="12" r="2.5"/><path d="M12 3v4M12 17v4M3 12h4M17 12h4"/></svg></span>
      <span>每日推荐</span>
    </router-link>
    <router-link to="/ranking" class="nav-item" active-class="active">
      <span class="nav-ic"><svg viewBox="0 0 24 24"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2Z"/></svg></span>
      <span>排行榜</span>
    </router-link>
    <router-link to="/fm" class="nav-item" active-class="active">
      <span class="nav-ic"><svg viewBox="0 0 24 24"><path d="M3 17.8 8.5 13l3.5 3 8.5-5.5"/><path d="M20 6v6M18 8h4"/></svg></span>
      <span>猜你喜欢</span>
    </router-link>

    <div class="nav-group">我的音乐</div>
    <router-link to="/me" class="nav-item" active-class="active">
      <span class="nav-ic"><svg viewBox="0 0 24 24"><circle cx="12" cy="8" r="4"/><path d="M4 21c0-3.3 3.6-6 8-6s8 2.7 8 6"/></svg></span>
      <span>我的</span>
    </router-link>
    <router-link to="/recent" class="nav-item" active-class="active">
      <span class="nav-ic"><svg viewBox="0 0 24 24"><path d="M12 21v-8M12 13 6.5 7.5M12 13l5.5-5.5"/><path d="M4 5h16"/></svg></span>
      <span>最近播放</span>
    </router-link>
    <router-link to="/cloud" class="nav-item" active-class="active">
      <span class="nav-ic"><svg viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><path d="M7 10l5 5 5-5"/><path d="M12 15V3"/></svg></span>
      <span>音乐云盘</span>
    </router-link>
    <router-link to="/local" class="nav-item" active-class="active">
      <span class="nav-ic"><svg viewBox="0 0 24 24"><path d="M9 18V6l10-2v12"/><circle cx="6.5" cy="18" r="2.5"/><circle cx="16.5" cy="16" r="2.5"/></svg></span>
      <span>本地音乐</span>
    </router-link>
  </aside>
</template>

<script setup>
</script>

<style scoped>
.sidebar {
  background: var(--surface);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 22px 14px 18px;
  position: relative;
  z-index: 5;
}

.logo { display: flex; align-items: center; gap: 11px; padding: 2px 10px 22px; }
.logo-badge {
  width: 36px; height: 36px; border-radius: 11px;
  background: linear-gradient(135deg, #2CA6F8, #1E7FD4);
  display: grid; place-items: center;
  box-shadow: 0 4px 14px rgba(44,166,248,.4);
  overflow: hidden;
  flex-shrink: 0;
}
.logo-badge img { width: 100%; height: 100%; object-fit: cover; display: block; }
.logo-name { font-size: 17px; font-weight: 700; letter-spacing: -.02em; white-space: nowrap; }
.logo-name span { color: var(--primary); }

.nav-group {
  font-size: 11px; font-weight: 600; color: var(--text-3);
  letter-spacing: .1em; padding: 16px 12px 8px;
  text-transform: uppercase;
}
.nav-item {
  position: relative;
  display: flex; align-items: center; gap: 12px;
  padding: 8px 10px; margin: 2px 0;
  border-radius: var(--r-md);
  color: var(--text-2); font-size: 13.5px; font-weight: 500;
  cursor: pointer; user-select: none;
  transition: background .18s var(--ease), color .18s var(--ease);
}
.nav-ic {
  width: 28px; height: 28px; border-radius: 9px;
  display: grid; place-items: center; flex-shrink: 0;
  background: transparent;
  transition: background .18s var(--ease), color .18s var(--ease);
}
.nav-item svg { width: 17px; height: 17px; stroke: currentColor; fill: none; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
.nav-item:hover { background: var(--surface-3); color: var(--text-1); }
.nav-item.active {
  background: var(--primary-soft); color: var(--primary-deep); font-weight: 600;
}
.nav-item.active .nav-ic {
  background: var(--primary);
  color: #fff;
  box-shadow: 0 3px 10px rgba(44,166,248,.4);
}
.nav-item.active::before {
  content: '';
  position: absolute; left: -14px; top: 50%; transform: translateY(-50%);
  width: 3px; height: 20px; border-radius: 99px;
  background: var(--primary);
}

.sidebar-spacer { flex: 1; }

@media (max-width: 860px) {
  .sidebar { display: none; }
}
</style>
