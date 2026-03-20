export default defineNuxtRouteMiddleware((to) => {
  const { isSignedIn } = useAuth()
  const localePath = useLocalePath()

  if (!isSignedIn.value) {
    return navigateTo({
      path: localePath('/sign-in'),
      query: { redirect: to.fullPath }
    })
  }
})
