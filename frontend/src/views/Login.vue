<template>
  <section class="page">
    <div class="panel">
      <h1>登录</h1>
      <p>登录页面骨架已创建，后续认证任务再接入真实接口。</p>
    </div>
  </section>
</template>

<style scoped>
.page {
  display: grid;
  min-height: calc(100vh - 56px);
  place-items: center;
  padding: 24px;
}

.panel {
  width: min(420px, 100%);
  padding: 24px;
  border: 1px solid #dde3ee;
  border-radius: 8px;
  background: #ffffff;
}

h1 {
  margin: 0 0 12px;
  font-size: 24px;
}

p {
  margin: 0;
  color: #667085;
  line-height: 1.7;
}
</style>
