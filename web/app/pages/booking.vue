<script setup>
import { ref, computed, watch, onMounted } from 'vue'

definePageMeta({
  middleware: 'auth'
})

const { t } = useI18n()
const route = useRoute()

useHead({
  title: `${t('booking.title')} | VTC.solutions`
})

// ---------------------------------------------------------------------------
// Address autocomplete
// ---------------------------------------------------------------------------

const ADRESSE_API = 'https://api-adresse.data.gouv.fr/search/'

const pickup = ref({
  street: '',
  zip: '',
  city: '',
  extra: '',
  coords: null // [lat, lng]
})

const dropoff = ref({
  street: '',
  zip: '',
  city: '',
  extra: '',
  coords: null
})

const pickupSuggestions = ref([])
const dropoffSuggestions = ref([])
const pickupLoading = ref(false)
const dropoffLoading = ref(false)

async function fetchSuggestions(query, targetRef, loadingRef) {
  const q = [query.street, query.zip, query.city].filter(Boolean).join(' ').trim()
  if (q.length < 3) {
    targetRef.value = []
    return
  }
  loadingRef.value = true
  try {
    const res = await fetch(`${ADRESSE_API}?q=${encodeURIComponent(q)}&limit=5&countrycodes=fr`)
    const data = await res.json()
    targetRef.value = data.features || []
  } catch {
    targetRef.value = []
  } finally {
    loadingRef.value = false
  }
}

function selectSuggestion(feature, targetAddress, suggestionsRef) {
  const p = feature.properties
  targetAddress.street = p.name || ''
  targetAddress.zip = p.postcode || ''
  targetAddress.city = p.city || ''
  targetAddress.coords = [feature.geometry.coordinates[1], feature.geometry.coordinates[0]]
  suggestionsRef.value = []
}

let pickupTimer = null
let dropoffTimer = null

function onPickupInput() {
  clearTimeout(pickupTimer)
  pickupTimer = setTimeout(() => fetchSuggestions(pickup.value, pickupSuggestions, pickupLoading), 300)
}

function onDropoffInput() {
  clearTimeout(dropoffTimer)
  dropoffTimer = setTimeout(() => fetchSuggestions(dropoff.value, dropoffSuggestions, dropoffLoading), 300)
}

// ---------------------------------------------------------------------------
// Pre-fill from query params (landing page redirects)
// ---------------------------------------------------------------------------

onMounted(() => {
  if (route.query.pickup) pickup.value.street = route.query.pickup
  if (route.query.dropoff) dropoff.value.street = route.query.dropoff
})

// ---------------------------------------------------------------------------
// Date & time
// ---------------------------------------------------------------------------

const today = new Date().toISOString().split('T')[0]
const bookingDate = ref('')
const bookingTime = ref('')

// ---------------------------------------------------------------------------
// Passengers & luggage
// ---------------------------------------------------------------------------

const passengers = ref(1)
const suitcases = ref(0)

const MAX_PASSENGERS = 6

function clamp(val, min, max) {
  return Math.min(Math.max(val, min), max)
}

// ---------------------------------------------------------------------------
// Special items
// ---------------------------------------------------------------------------

const specialItems = ref([
  { id: 'surfboard',   label: 'booking.items.surfboard',   supplement: 5,  selected: false },
  { id: 'bicycle',     label: 'booking.items.bicycle',     supplement: 5,  selected: false },
  { id: 'wheelchair',  label: 'booking.items.wheelchair',  supplement: 0,  selected: false },
  { id: 'golf',        label: 'booking.items.golf',        supplement: 5,  selected: false },
  { id: 'ski',         label: 'booking.items.ski',         supplement: 5,  selected: false },
  { id: 'pet',         label: 'booking.items.pet',         supplement: 0,  selected: false },
])

function toggleItem(item) {
  item.selected = !item.selected
}

// ---------------------------------------------------------------------------
// Fare estimate (flat rate per km, supplements)
// ---------------------------------------------------------------------------

const BASE_RATE = 1.8   // €/km
const BASE_FARE = 8     // minimum €
const routeDistance = ref(null) // km, set by map component

const supplementTotal = computed(() =>
  specialItems.value.filter(i => i.selected).reduce((sum, i) => sum + i.supplement, 0)
)

const estimatedFare = computed(() => {
  if (!routeDistance.value) return null
  const dist = Math.round(BASE_FARE + routeDistance.value * BASE_RATE + supplementTotal.value)
  return dist
})

const selectedSupplements = computed(() =>
  specialItems.value.filter(i => i.selected && i.supplement > 0)
)

// ---------------------------------------------------------------------------
// Map
// ---------------------------------------------------------------------------

const mapCenter = ref([43.39, -1.66]) // Pays Basque default
const mapZoom = ref(10)
const routeCoords = ref([])

watch([() => pickup.value.coords, () => dropoff.value.coords], async ([pCoords, dCoords]) => {
  if (!pCoords || !dCoords) {
    routeCoords.value = []
    routeDistance.value = null
    return
  }
  await fetchRoute(pCoords, dCoords)
})

async function fetchRoute(from, to) {
  try {
    const url = `https://router.project-osrm.org/route/v1/driving/${from[1]},${from[0]};${to[1]},${to[0]}?overview=full&geometries=geojson`
    const res = await fetch(url)
    const data = await res.json()
    if (data.routes && data.routes[0]) {
      const coords = data.routes[0].geometry.coordinates.map(([lng, lat]) => [lat, lng])
      routeCoords.value = coords
      routeDistance.value = Math.round(data.routes[0].distance / 1000)
      // fit map between both points
      mapCenter.value = [
        (from[0] + to[0]) / 2,
        (from[1] + to[1]) / 2
      ]
      mapZoom.value = 10
    }
  } catch {
    routeCoords.value = []
  }
}

// ---------------------------------------------------------------------------
// Notes
// ---------------------------------------------------------------------------

const notes = ref('')

// ---------------------------------------------------------------------------
// Submit
// ---------------------------------------------------------------------------

const submitting = ref(false)
const submitted = ref(false)
const submitError = ref('')

const formValid = computed(() =>
  pickup.value.coords &&
  dropoff.value.coords &&
  bookingDate.value &&
  bookingTime.value
)

async function submitBooking() {
  if (!formValid.value) return
  submitting.value = true
  submitError.value = ''
  try {
    const payload = {
      pickup: {
        street: pickup.value.street,
        zip: pickup.value.zip,
        city: pickup.value.city,
        extra: pickup.value.extra,
        lat: pickup.value.coords[0],
        lng: pickup.value.coords[1]
      },
      dropoff: {
        street: dropoff.value.street,
        zip: dropoff.value.zip,
        city: dropoff.value.city,
        extra: dropoff.value.extra,
        lat: dropoff.value.coords[0],
        lng: dropoff.value.coords[1]
      },
      date: bookingDate.value,
      time: bookingTime.value,
      passengers: passengers.value,
      suitcases: suitcases.value,
      special_items: specialItems.value.filter(i => i.selected).map(i => i.id),
      notes: notes.value,
      estimated_fare: estimatedFare.value,
      distance_km: routeDistance.value
    }
    const res = await fetch('/api/bookings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    submitted.value = true
  } catch (e) {
    submitError.value = e.message
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <UContainer class="py-8">

    <!-- Success state -->
    <div v-if="submitted" class="flex flex-col items-center justify-center py-24 gap-4">
      <UIcon name="i-lucide-circle-check" class="text-primary w-16 h-16" />
      <h2 class="text-2xl font-semibold">{{ $t('booking.confirmed_title') }}</h2>
      <p class="text-muted text-center max-w-sm">{{ $t('booking.confirmed_desc') }}</p>
      <UButton :to="localePath('/')" variant="subtle" label="Back to home" />
    </div>

    <!-- Booking form -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-[480px_1fr] gap-0 border border-default rounded-xl overflow-hidden min-h-[700px]">

      <!-- ----------------------------------------------------------------- -->
      <!-- Left panel — form                                                  -->
      <!-- ----------------------------------------------------------------- -->
      <div class="flex flex-col divide-y divide-default overflow-y-auto max-h-[85vh] lg:max-h-none">

        <!-- Pickup -->
        <div class="p-5 flex flex-col gap-3">
          <p class="text-xs text-muted uppercase tracking-widest font-medium">
            {{ $t('booking.pickup') }}
          </p>

          <!-- Street + ZIP -->
          <div class="relative flex gap-2">
            <div class="flex-1 relative">
              <UInput
                v-model="pickup.street"
                :placeholder="$t('booking.street')"
                icon="i-lucide-map-pin"
                class="w-full"
                @input="onPickupInput"
              />
              <!-- Autocomplete dropdown -->
              <div
                v-if="pickupSuggestions.length"
                class="absolute z-50 top-full mt-1 w-full bg-default border border-default rounded-lg shadow-lg overflow-hidden"
              >
                <button
                  v-for="s in pickupSuggestions"
                  :key="s.properties.id"
                  class="w-full text-left px-3 py-2 text-sm hover:bg-elevated transition-colors"
                  @click="selectSuggestion(s, pickup, pickupSuggestions)"
                >
                  <span class="font-medium">{{ s.properties.name }}</span>
                  <span class="text-muted ml-1">{{ s.properties.postcode }} {{ s.properties.city }}</span>
                </button>
              </div>
            </div>
            <UInput
              v-model="pickup.zip"
              :placeholder="$t('booking.zip')"
              class="w-24"
              @input="onPickupInput"
            />
          </div>

          <UInput
            v-model="pickup.city"
            :placeholder="$t('booking.city')"
            icon="i-lucide-building-2"
            @input="onPickupInput"
          />

          <UInput
            v-model="pickup.extra"
            :placeholder="$t('booking.extra_pickup')"
            icon="i-lucide-info"
          />

          <p v-if="pickup.coords" class="text-xs text-primary flex items-center gap-1">
            <UIcon name="i-lucide-circle-check" class="w-3 h-3" />
            {{ $t('booking.address_confirmed') }}
          </p>
        </div>

        <!-- Dropoff -->
        <div class="p-5 flex flex-col gap-3">
          <p class="text-xs text-muted uppercase tracking-widest font-medium">
            {{ $t('booking.dropoff') }}
          </p>

          <div class="relative flex gap-2">
            <div class="flex-1 relative">
              <UInput
                v-model="dropoff.street"
                :placeholder="$t('booking.street')"
                icon="i-lucide-flag"
                class="w-full"
                @input="onDropoffInput"
              />
              <div
                v-if="dropoffSuggestions.length"
                class="absolute z-50 top-full mt-1 w-full bg-default border border-default rounded-lg shadow-lg overflow-hidden"
              >
                <button
                  v-for="s in dropoffSuggestions"
                  :key="s.properties.id"
                  class="w-full text-left px-3 py-2 text-sm hover:bg-elevated transition-colors"
                  @click="selectSuggestion(s, dropoff, dropoffSuggestions)"
                >
                  <span class="font-medium">{{ s.properties.name }}</span>
                  <span class="text-muted ml-1">{{ s.properties.postcode }} {{ s.properties.city }}</span>
                </button>
              </div>
            </div>
            <UInput
              v-model="dropoff.zip"
              :placeholder="$t('booking.zip')"
              class="w-24"
              @input="onDropoffInput"
            />
          </div>

          <UInput
            v-model="dropoff.city"
            :placeholder="$t('booking.city')"
            icon="i-lucide-building-2"
            @input="onDropoffInput"
          />

          <UInput
            v-model="dropoff.extra"
            :placeholder="$t('booking.extra_dropoff')"
            icon="i-lucide-info"
          />

          <p v-if="dropoff.coords" class="text-xs text-primary flex items-center gap-1">
            <UIcon name="i-lucide-circle-check" class="w-3 h-3" />
            {{ $t('booking.address_confirmed') }}
          </p>
        </div>

        <!-- Date & time -->
        <div class="p-5 flex flex-col gap-3">
          <p class="text-xs text-muted uppercase tracking-widest font-medium">
            {{ $t('booking.when') }}
          </p>
          <div class="grid grid-cols-2 gap-3">
            <UInput
              v-model="bookingDate"
              type="date"
              :min="today"
              icon="i-lucide-calendar"
            />
            <UInput
              v-model="bookingTime"
              type="time"
              icon="i-lucide-clock"
            />
          </div>
        </div>

        <!-- Passengers & luggage -->
        <div class="p-5 flex flex-col gap-4">
          <p class="text-xs text-muted uppercase tracking-widest font-medium">
            {{ $t('booking.passengers_luggage') }}
          </p>

          <div class="grid grid-cols-2 gap-4">
            <!-- Passengers -->
            <div>
              <p class="text-xs text-muted mb-2">{{ $t('booking.passengers') }}</p>
              <div class="flex items-center border border-default rounded-lg overflow-hidden">
                <button
                  class="px-3 py-2 bg-elevated hover:bg-accented transition-colors text-sm font-medium disabled:opacity-30"
                  :disabled="passengers <= 1"
                  @click="passengers = clamp(passengers - 1, 1, MAX_PASSENGERS)"
                >−</button>
                <span class="flex-1 text-center text-sm font-medium border-x border-default py-2">
                  {{ passengers }}
                </span>
                <button
                  class="px-3 py-2 bg-elevated hover:bg-accented transition-colors text-sm font-medium disabled:opacity-30"
                  :disabled="passengers >= MAX_PASSENGERS"
                  @click="passengers = clamp(passengers + 1, 1, MAX_PASSENGERS)"
                >+</button>
              </div>
              <p class="text-xs text-muted mt-1">{{ $t('booking.max_passengers') }}</p>
            </div>

            <!-- Suitcases -->
            <div>
              <p class="text-xs text-muted mb-2">{{ $t('booking.suitcases') }}</p>
              <div class="flex items-center border border-default rounded-lg overflow-hidden">
                <button
                  class="px-3 py-2 bg-elevated hover:bg-accented transition-colors text-sm font-medium disabled:opacity-30"
                  :disabled="suitcases <= 0"
                  @click="suitcases = clamp(suitcases - 1, 0, 6)"
                >−</button>
                <span class="flex-1 text-center text-sm font-medium border-x border-default py-2">
                  {{ suitcases }}
                </span>
                <button
                  class="px-3 py-2 bg-elevated hover:bg-accented transition-colors text-sm font-medium disabled:opacity-30"
                  :disabled="suitcases >= 6"
                  @click="suitcases = clamp(suitcases + 1, 0, 6)"
                >+</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Special items -->
        <div class="p-5 flex flex-col gap-3">
          <div>
            <p class="text-xs text-muted uppercase tracking-widest font-medium">
              {{ $t('booking.special_items') }}
            </p>
            <p class="text-xs text-muted mt-1">{{ $t('booking.special_items_note') }}</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="item in specialItems"
              :key="item.id"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-full border text-sm transition-colors"
              :class="item.selected
                ? 'border-primary bg-primary/10 text-primary font-medium'
                : 'border-default text-muted hover:border-muted'"
              @click="toggleItem(item)"
            >
              {{ $t(item.label) }}
              <span v-if="item.supplement > 0" class="text-xs opacity-70">+€{{ item.supplement }}</span>
            </button>
          </div>
        </div>

        <!-- Notes -->
        <div class="p-5 flex flex-col gap-3">
          <p class="text-xs text-muted uppercase tracking-widest font-medium">
            {{ $t('booking.notes') }}
          </p>
          <UTextarea
            v-model="notes"
            :placeholder="$t('booking.notes_placeholder')"
            :rows="3"
          />
        </div>

        <!-- Fare & submit -->
        <div class="p-5 mt-auto flex flex-col gap-3 sticky bottom-0 bg-default border-t border-default">
          <div class="flex justify-between items-start">
            <span class="text-sm text-muted">{{ $t('booking.estimated_fare') }}</span>
            <div class="text-right">
              <p v-if="estimatedFare" class="text-2xl font-semibold">€ {{ estimatedFare }}</p>
              <p v-else class="text-sm text-muted italic">{{ $t('booking.fare_pending') }}</p>
              <div v-if="selectedSupplements.length" class="flex flex-col items-end gap-0.5 mt-1">
                <p
                  v-for="s in selectedSupplements"
                  :key="s.id"
                  class="text-xs text-muted"
                >
                  + €{{ s.supplement }} {{ $t(s.label) }}
                </p>
              </div>
            </div>
          </div>

          <p v-if="submitError" class="text-xs text-red-500">{{ submitError }}</p>

          <UButton
            block
            size="lg"
            :disabled="!formValid || submitting"
            :loading="submitting"
            :label="$t('booking.confirm')"
            @click="submitBooking"
          />

          <p v-if="!formValid" class="text-xs text-center text-muted">
            {{ $t('booking.fill_required') }}
          </p>
        </div>

      </div>

      <!-- ----------------------------------------------------------------- -->
      <!-- Right panel — map                                                  -->
      <!-- ----------------------------------------------------------------- -->
      <div class="hidden lg:block relative min-h-[500px] bg-elevated">
        <NuxtLeaflet
          :center="mapCenter"
          :zoom="mapZoom"
          class="w-full h-full min-h-[700px]"
        >
          <LTileLayer
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          />

          <!-- Pickup marker -->
          <LMarker v-if="pickup.coords" :lat-lng="pickup.coords">
            <LPopup>{{ pickup.street }}, {{ pickup.city }}</LPopup>
          </LMarker>

          <!-- Dropoff marker -->
          <LMarker v-if="dropoff.coords" :lat-lng="dropoff.coords">
            <LPopup>{{ dropoff.street }}, {{ dropoff.city }}</LPopup>
          </LMarker>

          <!-- Route polyline -->
          <LPolyline
            v-if="routeCoords.length"
            :lat-lngs="routeCoords"
            color="#00A155"
            :weight="4"
            :opacity="0.8"
          />
        </NuxtLeaflet>

        <!-- Distance badge -->
        <div
          v-if="routeDistance"
          class="absolute bottom-4 right-4 z-[1000] bg-default border border-default rounded-lg px-3 py-2 text-sm font-medium shadow"
        >
          {{ routeDistance }} km
        </div>

        <!-- Placeholder when no addresses yet -->
        <div
          v-if="!pickup.coords && !dropoff.coords"
          class="absolute inset-0 z-[1000] flex items-center justify-center pointer-events-none"
        >
          <p class="text-sm text-muted bg-default/80 px-4 py-2 rounded-lg border border-default">
            {{ $t('booking.map_hint') }}
          </p>
        </div>
      </div>

    </div>
  </UContainer>
</template>
