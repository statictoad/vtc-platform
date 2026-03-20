// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@nuxt/eslint',
    '@clerk/nuxt',
    '@nuxt/ui',
    '@nuxtjs/i18n',
    '@nuxtjs/sitemap',
    '@nuxtjs/robots'
  ],

  devtools: {
    enabled: true
  },

  css: ['~/assets/css/main.css'],

  routeRules: {
    '/': { prerender: true }
  },

  compatibilityDate: '2025-01-15',

  nitro: {
    devProxy: {
      '/api': {
        // This removes the dependency on the 'process' type
        // by looking it up via the globalThis or a simple string
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  },

  i18n: {
    langDir: 'locales/',
    locales: [
      { code: 'fr', iso: 'fr-FR', name: 'Français', file: 'fr.json' },
      { code: 'en', iso: 'en-GB', name: 'English', file: 'en.json' },
      { code: 'ja', iso: 'ja-JP', name: '日本語', file: 'ja.json' }
    ],
    defaultLocale: 'en',
    strategy: 'prefix',
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: 'i18n_redirected',
      redirectOn: 'root'
    }
  },

  robots: {
    disallow: ['/api', '/admin'], // Keep Go backend routes private
    allow: '/',
    sitemap: 'https://vtc.solutions/sitemap.xml'
  },

  sitemap: {
    autoLastmod: true // This ensures /fr/, /en/, etc. are all included
  },

  vite: {
    optimizeDeps: {
      include: [
        '@clerk/vue',
        '@vue/devtools-core',
        '@vue/devtools-kit'
      ]
    }
  }
})
