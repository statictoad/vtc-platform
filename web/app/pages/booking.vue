<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

const { t } = useI18n()
const localePath = useLocalePath()
const route = useRoute()

useHead({
  title: `${t('booking.title')} | VTC.solutions`
})

// ---------------------------------------------------------------------------
// Addresses
// ---------------------------------------------------------------------------

const pickup = ref({ street: '', zip: '', city: '', extra: '', coords: null })
const dropoff = ref({ street: '', zip: '', city: '', extra: '', coords: null })

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

// ---------------------------------------------------------------------------
// Special items
// ---------------------------------------------------------------------------

const specialItems = ref([
  { id: 'surfboard', label: 'booking.items.surfboard', supplement: 5, selected: false },
  { id: 'bicycle', label: 'booking.items.bicycle', supplement: 5, selected: false },
  { id: 'wheelchair', label: 'booking.items.wheelchair', supplement: 0, selected: false },
  { id: 'golf', label: 'booking.items.golf', supplement: 0, selected: false },
  { id: 'ski', label: 'booking.items.ski', supplement: 0, selected: false },
  { id: 'pet', label: 'booking.items.pet', supplement: 0, selected: false }
])

function toggleItem(item) {
  item.selected = !item.selected
}

// ---------------------------------------------------------------------------
// Fare estimate
// ---------------------------------------------------------------------------

const BASE_RATE = 1.8
const BASE_FARE = 8
const routeDistance = ref(null)

const supplementTotal = computed(() =>
  specialItems.value.filter(i => i.selected).reduce((sum, i) => sum + i.supplement, 0)
)

const estimatedFare = computed(() => {
  if (!routeDistance.value) return null
  return Math.round(BASE_FARE + routeDistance.value * BASE_RATE + supplementTotal.value)
})

const selectedSupplements = computed(() =>
  specialItems.value.filter(i => i.selected && i.supplement > 0)
)

// ---------------------------------------------------------------------------
// Map (vanilla Leaflet)
// ---------------------------------------------------------------------------

let map = null
let pickupMarker = null
let dropoffMarker = null
let routePolyline = null

async function initMap() {
  const L = (await import('leaflet')).default
  await import('leaflet/dist/leaflet.css')

  delete L.Icon.Default.prototype._getIconUrl
  L.Icon.Default.mergeOptions({
    iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
    iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
    shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png'
  })

  map = L.map('map').setView([43.39, -1.66], 10)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
  }).addTo(map)

  watch([() => pickup.value.coords, () => dropoff.value.coords], async ([pCoords, dCoords]) => {
    if (pCoords) {
      if (pickupMarker) pickupMarker.setLatLng(pCoords)
      else pickupMarker = L.marker(pCoords).addTo(map)
    } else if (pickupMarker) {
      map.removeLayer(pickupMarker)
      pickupMarker = null
    }

    if (dCoords) {
      if (dropoffMarker) dropoffMarker.setLatLng(dCoords)
      else dropoffMarker = L.marker(dCoords).addTo(map)
    } else if (dropoffMarker) {
      map.removeLayer(dropoffMarker)
      dropoffMarker = null
    }

    if (pCoords && dCoords) {
      await fetchRoute(L, pCoords, dCoords)
    } else {
      if (routePolyline) { map.removeLayer(routePolyline); routePolyline = null }
      routeDistance.value = null
    }
  })
}

async function fetchRoute(L, from, to) {
  try {
    const url = `https://router.project-osrm.org/route/v1/driving/${from[1]},${from[0]};${to[1]},${to[0]}?overview=full&geometries=geojson`
    const res = await fetch(url)
    const data = await res.json()
    if (data.routes?.[0]) {
      const coords = data.routes[0].geometry.coordinates.map(([lng, lat]) => [lat, lng])
      routeDistance.value = data.routes[0].distance
      if (routePolyline) map.removeLayer(routePolyline)
      routePolyline = L.polyline(coords, { color: '#00A155', weight: 4, opacity: 0.8 }).addTo(map)
      map.fitBounds(routePolyline.getBounds(), { padding: [40, 40] })
    }
  } catch {
    routeDistance.value = null
  }
}

onMounted(() => {
  if (route.query.pickup) pickup.value.street = route.query.pickup
  if (route.query.dropoff) dropoff.value.street = route.query.dropoff
  initMap()
})

onUnmounted(() => {
  if (map) { map.remove(); map = null }
})

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
  pickup.value.coords && dropoff.value.coords && bookingDate.value && bookingTime.value
)

async function submitBooking() {
  if (!formValid.value) return
  submitting.value = true
  submitError.value = ''
  try {
    const payload = {
      pickup: { ...pickup.value, lat: pickup.value.coords[0], lng: pickup.value.coords[1] },
      dropoff: { ...dropoff.value, lat: dropoff.value.coords[0], lng: dropoff.value.coords[1] },
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
    <div
      v-if="submitted"
      class="flex flex-col items-center justify-center py-24 gap-4"
    >
      <UIcon
        name="i-lucide-circle-check"
        class="text-primary w-16 h-16"
      />
      <h2 class="text-2xl font-semibold">
        {{ $t('booking.confirmed_title') }}
      </h2>
      <p class="text-muted text-center max-w-sm">
        {{ $t('booking.confirmed_desc') }}
      </p>
      <UButton
        :to="localePath('/')"
        variant="subtle"
        :label="$t('booking.back_home')"
      />
    </div>

    <!-- Booking form -->
    <div
      v-else
      class="grid grid-cols-1 lg:grid-cols-[480px_1fr] border border-default rounded-xl overflow-hidden min-h-[700px]"
    >
      <!-- Left panel — form -->
      <div class="flex flex-col divide-y divide-default overflow-y-auto max-h-[85vh] lg:max-h-none">
        <!-- Pickup -->
        <UFormField
          :label="$t('booking.pickup')"
          class="p-5"
        >
          <BookingAddressBlock
            v-model="pickup"
            icon="i-lucide-map-pin"
            :extra-placeholder="$t('booking.extra_pickup')"
          />
        </UFormField>

        <!-- Dropoff -->
        <UFormField
          :label="$t('booking.dropoff')"
          class="p-5"
        >
          <BookingAddressBlock
            v-model="dropoff"
            icon="i-lucide-flag"
            :extra-placeholder="$t('booking.extra_dropoff')"
          />
        </UFormField>

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
            <BookingCounter
              v-model="passengers"
              :min="1"
              :max="6"
              :label="$t('booking.passengers')"
              :hint="$t('booking.max_passengers')"
            />
            <BookingCounter
              v-model="suitcases"
              :min="0"
              :max="6"
              :label="$t('booking.suitcases')"
            />
          </div>
        </div>

        <!-- Special items -->
        <div class="p-5 flex flex-col gap-3">
          <div>
            <p class="text-xs text-muted uppercase tracking-widest font-medium">
              {{ $t('booking.special_items') }}
            </p>
            <p class="text-xs text-muted mt-1">
              {{ $t('booking.special_items_note') }}
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UButton
              v-for="item in specialItems"
              :key="item.id"
              size="sm"
              :variant="item.selected ? 'soft' : 'outline'"
              :color="item.selected ? 'primary' : 'neutral'"
              :label="item.supplement > 0 ? `${$t(item.label)} +€${item.supplement}` : $t(item.label)"
              class="rounded-full"
              @click="toggleItem(item)"
            />
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
              <p
                v-if="estimatedFare"
                class="text-2xl font-semibold"
              >
                € {{ estimatedFare }}
              </p>
              <p
                v-else
                class="text-sm text-muted italic"
              >
                {{ $t('booking.fare_pending') }}
              </p>
              <div
                v-if="selectedSupplements.length"
                class="flex flex-col items-end gap-0.5 mt-1"
              >
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

          <UAlert
            v-if="submitError"
            color="error"
            variant="soft"
            :description="submitError"
          />

          <UButton
            block
            size="lg"
            :disabled="!formValid || submitting"
            :loading="submitting"
            :label="$t('booking.confirm')"
            @click="submitBooking"
          />

          <p
            v-if="!formValid"
            class="text-xs text-center text-muted"
          >
            {{ $t('booking.fill_required') }}
          </p>
        </div>
      </div>

      <!-- Right panel — map -->
      <div class="hidden lg:block relative min-h-[500px] bg-elevated">
        <div
          id="map"
          class="w-full h-full min-h-[700px]"
        />

        <UBadge
          v-if="routeDistance"
          class="absolute bottom-4 right-4 z-[1000]"
          color="neutral"
          variant="soft"
          size="lg"
        >
          {{ routeDistance }} km
        </UBadge>

        <div
          v-if="!pickup.coords && !dropoff.coords"
          class="absolute inset-0 z-[400] flex items-center justify-center pointer-events-none"
        >
          <UBadge
            color="neutral"
            variant="soft"
            size="lg"
          >
            {{ $t('booking.map_hint') }}
          </UBadge>
        </div>
      </div>
    </div>
  </UContainer>
</template>
