<script setup>
const props = defineProps({
  modelValue: {
    type: Number,
    required: true
  },
  min: {
    type: Number,
    default: 0
  },
  max: {
    type: Number,
    default: 10
  },
  label: {
    type: String,
    default: ''
  },
  hint: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue'])

function decrement() {
  if (props.modelValue > props.min) emit('update:modelValue', props.modelValue - 1)
}

function increment() {
  if (props.modelValue < props.max) emit('update:modelValue', props.modelValue + 1)
}
</script>

<template>
  <div class="flex flex-col gap-1.5">
    <p
      v-if="label"
      class="text-xs text-muted"
    >
      {{ label }}
    </p>
    <UFieldGroup class="w-full">
      <UButton
        icon="i-lucide-minus"
        color="neutral"
        variant="outline"
        :disabled="modelValue <= min"
        @click="decrement"
      />
      <UInput
        :model-value="modelValue"
        class="flex-1 text-center"
        readonly
      />
      <UButton
        icon="i-lucide-plus"
        color="neutral"
        variant="outline"
        :disabled="modelValue >= max"
        @click="increment"
      />
    </UFieldGroup>
    <p
      v-if="hint"
      class="text-xs text-muted"
    >
      {{ hint }}
    </p>
  </div>
</template>
