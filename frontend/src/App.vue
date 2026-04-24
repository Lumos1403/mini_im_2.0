<template>
  <main class="app-shell">
    <nav class="app-nav">
      <RouterLink to="/login">登录</RouterLink>
      <RouterLink to="/register">注册</RouterLink>
      <RouterLink to="/chat">聊天</RouterLink>
    </nav>
    <RouterView />
  </main>
</template>

<style scoped>
.app-shell {
  min-height: 100vh;
  color: #172033;
  background: #f6f8fb;
  font-family:
    Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI",
    sans-serif;
}

.app-nav {
  display: flex;
  gap: 16px;
  align-items: center;
  height: 56px;
  padding: 0 24px;
  border-bottom: 1px solid #dde3ee;
  background: #ffffff;
}

.app-nav a {
  color: #4a5568;
  text-decoration: none;
}

.app-nav a.router-link-active {
  color: #2563eb;
  font-weight: 600;
}
</style>
