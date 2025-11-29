<script setup lang="ts">
import { Role } from "@/components/types.ts";
import RoomForm from "./RoomForm.vue";

type Props = {
  activatorText: string;
  cardTitle: string;
  showPasswordInput?: boolean;
  errorMessage?: string;
};
const props = withDefaults(defineProps<Props>(), {
  showPasswordInput: false,
  errorMessage: undefined,
});
const emits = defineEmits<{
  (e: "submit"): void;
}>();
const name = defineModel("name", { required: true, type: String, default: "" });
const role = defineModel("role", { required: true, type: String, default: Role.Empty });
const password = defineModel("password", { type: String, default: "", required: false });
const showDialog = defineModel("showDialog", { type: Boolean, default: false, required: false });
</script>

<template>
  <v-dialog v-model="showDialog" width="500">
    <template #activator="{ props: activatorProps }">
      <v-btn v-bind="activatorProps" color="primary" :text="props.activatorText" />
    </template>

    <room-form
      v-model:name="name"
      v-model:role="role"
      v-model:password="password"
      :error-message="props.errorMessage"
      :show-password-input="props.showPasswordInput"
      :title="cardTitle"
      @submit="emits('submit')"
    />
  </v-dialog>
</template>

<style scoped></style>
