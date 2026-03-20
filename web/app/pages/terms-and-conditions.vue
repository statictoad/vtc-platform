<script setup>
const { t } = useI18n()

useHead({
  title: `${t('terms.title')} | VTC.solutions`,
  meta: [
    { name: 'robots', content: 'noindex, follow' } // Standard for legal pages
  ]
})
</script>

<template>
  <UContainer class="py-12">
    <section class="prose dark:prose-invert max-w-none">
      <h1 class="text-3xl font-bold mb-12">
        {{ $t('terms.title') }}
      </h1>

      <div class="space-y-12">
        <div class="border-l-2 border-gray-100 dark:border-gray-800 pl-6">
          <h2 class="text-xl font-semibold mb-4">
            1. {{ $t('terms.legal_id') }}
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <p><strong>{{ $t('terms.company_name') }}:</strong> [Your Name/Company]</p>
            <p><strong>SIRET:</strong> [Your Number]</p>
            <p><strong>Numéro REVTC:</strong> [EVTC Number]</p>
            <p><strong>{{ $t('terms.address') }}:</strong> [Your Address, France]</p>
          </div>
        </div>

        <div class="border-l-2 border-gray-100 dark:border-gray-800 pl-6">
          <h2 class="text-xl font-semibold mb-2">
            2. {{ $t('terms.service_title') }}
          </h2>
          <p>{{ $t('terms.service_desc') }}</p>
        </div>

        <div class="border-l-2 border-gray-100 dark:border-gray-800 pl-6">
          <h2 class="text-xl font-semibold mb-2">
            3. {{ $t('terms.pricing_title') }}
          </h2>
          <p>{{ $t('terms.pricing_desc') }}</p>
        </div>

        <div class="border-l-2 border-gray-100 dark:border-gray-800 pl-6">
          <h2 class="text-xl font-semibold mb-2">
            4. {{ $t('terms.cancellation_title') }}
          </h2>
          <p>{{ $t('terms.cancellation_desc') }}</p>
        </div>

        <div class="border-l-2 border-gray-100 dark:border-gray-800 pl-6">
          <h2 class="text-xl font-semibold mb-2">
            5. {{ $t('terms.insurance_title') }}
          </h2>
          <div class="space-y-4">
            <p>{{ $t('terms.insurance_text') }}</p>
            <div class="p-4 bg-primary-50 dark:bg-primary-950/30 rounded-md border-l-4 border-primary text-sm italic">
              {{ $t('terms.insurance_details') }}
            </div>
          </div>
        </div>
      </div>
    </section>
  </UContainer>
</template>
