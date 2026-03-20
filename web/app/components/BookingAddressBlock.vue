<script setup>
const props = defineProps({
  modelValue: {
    type: Object,
    required: true
  },
  icon: {
    type: String,
    default: 'i-lucide-map-pin'
  },
  extraPlaceholder: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()

const ADDRESS_API = 'https://api-adresse.data.gouv.fr/search/'
const suggestions = ref([])
const loading = ref(false)
let suppress = false
let timer = null

const address = computed({
  get: () => props.modelValue,
  set: val => emit('update:modelValue', val)
})

async function fetchSuggestions() {
  const q = [address.value.street, address.value.zip, address.value.city].filter(Boolean).join(' ').trim()
  if (q.length < 3) {
    suggestions.value = []
    return
  }
  loading.value = true
  try {
    const res = await fetch(`${ADDRESS_API}?q=${encodeURIComponent(q)}&limit=5&countrycodes=fr`)
    const data = await res.json()
    suggestions.value = data.features || []
  } catch {
    suggestions.value = []
  } finally {
    loading.value = false
  }
}

function onInput() {
  if (suppress) {
    suppress = false
    return
  }
  clearTimeout(timer)
  timer = setTimeout(fetchSuggestions, 300)
}

function selectSuggestion(feature) {
  const p = feature.properties
  emit('update:modelValue', {
    ...props.modelValue,
    street: p.name || '',
    zip: p.postcode || '',
    city: p.city || '',
    coords: [feature.geometry.coordinates[1], feature.geometry.coordinates[0]]
  })
  suggestions.value = []
  suppress = true
}

function updateField(field, value) {
  emit('update:modelValue', { ...props.modelValue, [field]: value })
}
</script>

<template>
  <div class="flex flex-col gap-3">
    <div class="relative flex gap-2">
      <div class="flex-1 relative">
        <UInput
          :model-value="address.street"
          :placeholder="t('booking.street')"
          :icon="icon"
          :loading="loading"
          class="w-full"
          @update:model-value="val => updateField('street', val)"
          @input="onInput"
        />
        <div
          v-if="suggestions.length"
          class="absolute z-50 top-full mt-1 w-full bg-default border border-default rounded-lg overflow-hidden shadow-lg"
        >
          <button
            v-for="s in suggestions"
            :key="s.properties.id"
            class="w-full text-left px-3 py-2 text-sm hover:bg-elevated transition-colors"
            @click="selectSuggestion(s)"
          >
            <span class="font-medium">{{ s.properties.name }}</span>
            <span class="text-muted ml-1">{{ s.properties.postcode }} {{ s.properties.city }}</span>
          </button>
        </div>
      </div>
      <UInput
        :model-value="address.zip"
        :placeholder="t('booking.zip')"
        class="w-24"
        @update:model-value="val => updateField('zip', val)"
        @input="onInput"
      />
    </div>

    <UInput
      :model-value="address.city"
      :placeholder="t('booking.city')"
      icon="i-lucide-building-2"
      @update:model-value="val => updateField('city', val)"
      @input="onInput"
    />

    <UInput
      :model-value="address.extra"
      :placeholder="extraPlaceholder"
      icon="i-lucide-info"
      @update:model-value="val => updateField('extra', val)"
    />

    <p
      v-if="address.coords"
      class="text-xs text-primary flex items-center gap-1"
    >
      <UIcon name="i-lucide-circle-check" class="w-3 h-3" />
      {{ t('booking.address_confirmed') }}
    </p>
  </div>
</template>
