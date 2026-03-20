<template>
  <div>
    <UPageHero
      :title="$t('hero_title')"
      description="From the bay of Saint-Jean-de-Luz to the surf spots of Hossegor. Reliable service, affordable prices."
      :links="[{
        label: 'Book now',
        to: 'https://ui.nuxt.com/docs/getting-started/installation/nuxt',
        target: '_blank',
        trailingIcon: 'i-ic:sharp-near-me',
        size: 'xl'
      }, {
        label: 'See prices',
        to: 'https://github.com/nuxt-ui-templates/starter',
        target: '_blank',
        icon: 'i-ic:baseline-euro',
        size: 'xl',
        color: 'neutral',
        variant: 'subtle'
      }]"
    />

    <UPageSection
      id="features"
      title="Tailored for the Coast"
      description="Everything you need for a smooth journey in the South West."
      :features="[{
        icon: 'i-lucide-users',
        title: 'Up to 6 Passengers',
        description: 'Travel together in comfort. Note: Luggage capacity depends on the number of passengers.'
      }, {
        icon: 'i-lucide-waves', // Changed icon to waves/surf
        title: 'Surf & Bike Friendly',
        description: 'Specialized racks available for surfboards and bicycles (pre-order required).'
      }, {
        icon: 'i-lucide-map-pinned',
        title: 'Local & Long Distance',
        description: 'Serving the Basque Coast, Sud Landes, and any destination across France or Spain.'
      }, {
        icon: 'i-lucide-languages', // Changed icon to languages
        title: 'Bilingual Driver',
        description: 'Full English and French fluency to ensure smooth communication for international travelers.'
      }, {
        icon: 'i-lucide-accessibility', // Changed icon to accessibility
        title: 'Inclusive Service',
        description: 'Accommodating passengers with reduced mobility and foldable wheelchairs. Assistance provided.'
      }, {
        icon: 'i-lucide-plane-takeoff', // Changed icon to plane
        title: 'Hub Connections',
        description: 'Reliable transfers to Biarritz, San Sebastián, Bilbao, and Bordeaux airports or stations.'
      }]"
    />

    <UPageSection>
      <UPageCTA
        title="Need a custom itinerary?"
        description="For weddings, events, or specific sporting needs, please contact us directly."
        variant="subtle"
      >
        <ClientOnly>
          <div class="flex flex-col items-center gap-6">
            <div class="flex flex-wrap justify-center gap-4">
              <UButton
                label="Email"
                icon="i-lucide-mail"
                color="neutral"
                variant="subtle"
                to="mailto:contact@vtc.solutions"
                size="lg"
              />
              <UButton
                label="WhatsApp"
                icon="i-simple-icons-whatsapp"
                color="neutral"
                variant="subtle"
                to="https://wa.me/yournumber"
                size="lg"
              />
              <UButton
                label="Signal"
                icon="i-simple-icons-signal"
                color="neutral"
                variant="subtle"
                to="https://signal.me/#p/yournumber"
                size="lg"
              />
              <UButton
                label="Telegram"
                icon="i-simple-icons-telegram"
                color="neutral"
                variant="subtle"
                to="https://t.me/yourusername"
                size="lg"
              />
            </div>
            <p class="text-sm text-center leading-relaxed">
              To receive an accurate quote, please provide as many details as possible: date, time, passenger and luggage counts, etc.
              Response times vary depending on the driver's current driving assignments.
            </p>
          </div>
        </ClientOnly>
      </UPageCTA>
    </UPageSection>
  </div>
</template>
