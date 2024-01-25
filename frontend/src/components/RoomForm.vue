<script setup lang="ts">
import { Role } from "@/components/types";
import { computed } from "vue";

type Props = {
  name: string;
  roomId: string;
  role: Role;
  errorMessage: string | null;
};

const props = withDefaults(defineProps<Props>(), {
  errorMessage: null,
});
const emit = defineEmits<{
  (e: "update:name", value: string): void;
  (e: "update:roomId", value: string): void;
  (e: "update:role", value: Role): void;
  (e: "submit"): void;
}>();

const isButtonEnabled = computed(
  () => props.roomId !== "" && props.name !== "" && props.role !== "",
);

const textFieldRules = computed(() => [
  (value: string) => !!value || "Fehler: Hier müsste eigentlich was stehen",
]);
</script>

<template>
  <v-form :fast-fail="true" @submit.prevent="emit('submit')" validate-on="input">
    <v-col>
      <v-text-field
        label="Raum"
        :model-value="props.roomId"
        @update:modelValue="emit('update:roomId', $event)"
        required
        :rules="textFieldRules"
      />
      <v-text-field
        label="Name"
        :model-value="props.name"
        @update:modelValue="emit('update:name', $event)"
        required
        :rules="textFieldRules"
      />
    </v-col>

    <v-radio-group
      label="Deine Rolle"
      :model-value="props.role"
      @update:modelValue="emit('update:role', $event)"
    >
      <v-radio label="Product Owner" :value="Role.ProductOwner"></v-radio>
      <v-radio label="Entwickler" :value="Role.Developer"></v-radio>
    </v-radio-group>

    <v-col v-if="errorMessage !== ''">
      <v-alert color="error" :text="errorMessage" />
    </v-col>

    <v-col class="text-right">
      <v-btn
        type="submit"
        color="primary"
        prepend-icon="mdi-connection"
        :disabled="!isButtonEnabled"
        >Verbinden</v-btn
      >
    </v-col>
  </v-form>
</template>

<style scoped></style>
