<script setup>
const { handleRedirectCallback } = useSignIn()
await handleRedirectCallback()
</script>

<template>
  <div class="flex items-center justify-center min-h-screen">
    <UIcon
      name="i-lucide-loader"
      class="animate-spin w-8 h-8 text-muted"
    />
  </div>
</template>
