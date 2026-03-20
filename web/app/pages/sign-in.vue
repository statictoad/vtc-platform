<script setup>
const route = useRoute()
const redirectUrl = computed(() => route.query.redirect || '/')
</script>

<template>
  <UContainer class="py-12">
    <div class="flex justify-center items-center">
      <SignIn :force-redirect-url="redirectUrl" />
    </div>
  </UContainer>
</template>
