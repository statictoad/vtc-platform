<script setup>
const route = useRoute()
const localePath = useLocalePath()

const redirectUrl = computed(() => {
  const redirect = route.query.redirect
  if (!redirect) {
    return localePath('/booking')
  }
  try {
    return new URL(redirect).pathname
  } catch {
    return redirect
  }
})
</script>

<template>
  <UContainer class="py-12">
    <div class="flex justify-center items-center">
      <SignIn :force-redirect-url="redirectUrl" />
    </div>
  </UContainer>
</template>
