<script setup lang="ts">
import { Role, RoundState } from "@/components/types";
import DeveloperCommandCenter from "@/components/DeveloperCommandCenter.vue";
import { computed } from "vue";
import ProductOwnerCommandCenter from "@/components/ProductOwnerCommandCenter.vue";

type Props = {
  userRole: Role;
  roundState: RoundState;
};

const props = defineProps<Props>();

const isDeveloper = computed(() => props.userRole === Role.Developer);
const isProductOwner = computed(() => props.userRole === Role.ProductOwner);
const roundIsInProgress = computed(() => props.roundState === RoundState.InProgress);
</script>

<template>
  <developer-command-center v-if="isDeveloper && roundIsInProgress" />
  <product-owner-command-center v-else-if="isProductOwner" />
</template>

<style scoped></style>
