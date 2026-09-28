<script setup lang="ts">
const props = defineProps<{
  value: string;
  reveal: boolean;
  estimated: boolean;
}>();
</script>

<template>
  <div :class="{ 'flip-card': true, hoverable: true }">
    <div :class="{ 'flip-card__inner': true, reveal: props.reveal }">
      <div
        :class="{
          'flip-card__front': true,
          'waiting-for-estimate': !props.estimated,
          estimated: props.estimated,
        }"
      />
      <div :class="{ 'flip-card__back': true }">
        <span v-if="props.reveal">
          <strong>{{ props.value }}</strong>
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.flip-card {
  width: 3rem;
  height: 5rem;
  text-align: center;
  user-select: none;
}

.flip-card__inner {
  position: relative;
  width: 100%;
  height: 100%;
  text-align: center;
  transition: transform 0.5s;
  transform-style: preserve-3d;
  border-radius: 5px;
}

.reveal {
  transform: rotateY(180deg);
}

.flip-card__front {
  background-color: gray;
}

.flip-card__front,
.flip-card__back {
  position: absolute;
  width: 100%;
  height: 100%;
  backface-visibility: hidden;
  border-radius: 5px;
}

.flip-card__back {
  display: flex;
  justify-content: center;
  align-items: center;
  transform: rotateY(180deg);
  border: 2px solid #2196f3;
  background-color: white;
}

.flip-card.hoverable:hover,
.flip-card.hoverable:focus-within {
  cursor: pointer;
  transform: translateY(-0.25rem);
}

.flip-card.hoverable:hover .flip-card__back {
  background-color: rgba(0, 0, 0, 0.05);
}

.waiting-for-estimate {
  background-color: gray;
}

.estimated {
  background-color: #2196f3;
}
</style>
