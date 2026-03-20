<script setup>
const { isSignedIn, isLoaded } = useAuth()
const switchLocalePath = useSwitchLocalePath()
const localePath = useLocalePath()
const title = $t('meta.default_title')
const description = $t('meta.default_description')

useHead({
  title: title,
  meta: [
    { name: 'viewport', content: 'width=device-width, initial-scale=1' }
  ],
  link: [
    { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }
  ]
})

useSeoMeta({
  description: description,
  ogTitle: title,
  ogDescription: description
})
</script>

<template>
  <UApp>
    <UHeader :ui="{ toggle: 'hidden' }">
      <template #left>
        <NuxtLink to="/">
          <AppLogo class="w-auto h-6 shrink-0" />
        </NuxtLink>
      </template>

      <template #right>
        <UColorModeButton class="portrait:hidden" />

        <ClientOnly>
          <ClerkLoaded>
            <template v-if="isLoaded && !isSignedIn">
              <SignUpButton mode="modal">
                <UButton
                  label="Register"
                  variant="ghost"
                />
              </SignUpButton>
              <SignInButton mode="modal">
                <UButton label="Connect" />
              </SignInButton>
            </template>

            <template v-if="isLoaded && isSignedIn">
              <UserButton after-sign-out-url="/" />
            </template>
          </ClerkLoaded>
        </ClientOnly>
      </template>
    </UHeader>

    <UMain>
      <NuxtPage />
    </UMain>

    <USeparator icon="i-lucide-car" />

    <UFooter class="pb-8">
      <template #left>
        <p class="text-sm text-muted">
          © {{ new Date().getFullYear() }} •
          <NuxtLink :to="localePath('/terms-and-conditions')">
            {{ $t('footer.terms') }}
          </NuxtLink>
        </p>
      </template>
      <template #right>
        <div class="flex flex-col gap-3">
          <div class="flex flex-wrap gap-x-4 gap-y-2 text-xs font-medium opacity-80">
            <NuxtLink
              :to="switchLocalePath('en')"
              class="hover:text-primary transition-colors"
            >
              English
            </NuxtLink>
            <NuxtLink
              :to="switchLocalePath('fr')"
              class="hover:text-primary transition-colors"
            >
              Français
            </NuxtLink>
            <NuxtLink
              :to="switchLocalePath('ja')"
              class="hover:text-primary transition-colors"
            >
              日本語
            </NuxtLink>
          </div>
        </div>
      </template>
    </UFooter>
  </UApp>
</template>
