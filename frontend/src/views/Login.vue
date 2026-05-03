<script setup lang="ts">
import AuthPanel from '../components/auth/AuthPanel.vue'
import AuthLayout from '../layouts/AuthLayout.vue'
</script>

<template>
  <AuthLayout>
    <AuthPanel initial-mode="login" />
  </AuthLayout>
</template>
