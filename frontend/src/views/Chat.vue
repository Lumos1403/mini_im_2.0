<template>
  <section class="chat-shell">
    <aside class="conversation-list">
      <h2>会话</h2>
      <p>会话列表占位</p>
    </aside>
    <section class="chat-main">
      <header class="chat-header">聊天窗口</header>
      <div class="message-area">消息列表占位</div>
      <footer class="composer">输入区占位</footer>
    </section>
  </section>
</template>

<style scoped>
.chat-shell {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  min-height: calc(100vh - 56px);
}

.conversation-list {
  padding: 20px;
  border-right: 1px solid #dde3ee;
  background: #ffffff;
}

.conversation-list h2 {
  margin: 0 0 12px;
  font-size: 18px;
}

.conversation-list p,
.message-area {
  color: #667085;
}

.chat-main {
  display: grid;
  grid-template-rows: 56px 1fr 88px;
  min-width: 0;
}

.chat-header,
.composer {
  display: flex;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid #dde3ee;
  background: #ffffff;
}

.message-area {
  display: grid;
  place-items: center;
  padding: 24px;
}

.composer {
  border-top: 1px solid #dde3ee;
  border-bottom: 0;
}

@media (max-width: 720px) {
  .chat-shell {
    grid-template-columns: 1fr;
  }

  .conversation-list {
    border-right: 0;
    border-bottom: 1px solid #dde3ee;
  }
}
</style>
