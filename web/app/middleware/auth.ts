export default defineNuxtRouteMiddleware(() => {
  const { isSignedIn } = useAuth()
  if (!isSignedIn.value) {
    return abortNavigation()
  }
})
