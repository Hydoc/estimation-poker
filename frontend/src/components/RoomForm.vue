<script setup lang="ts">
import { Role } from "@/components/types";
import { computed } from "vue";

type Props = {
  errorMessage?: string | null;
  isRoomIdDisabled?: boolean;
};

const name = defineModel("name", { required: true, default: "" });
const roomId = defineModel("roomId", { required: true, default: "" });
const role = defineModel("role", { required: true, default: Role.Empty });

const props = withDefaults(defineProps<Props>(), {
  errorMessage: null,
  isRoomIdDisabled: false,
});
const emit = defineEmits<{
  (e: "submit"): void;
}>();

const isButtonEnabled = computed(
  () => roomId.value !== "" && name.value !== "" && role.value !== "",
);

const textFieldRules = computed(() => [
  (value: string) => !!value || "Fehler: Hier müsste eigentlich was stehen",
]);
</script>

<template>
  <v-form :fast-fail="true" @submit.prevent="emit('submit')" validate-on="input">
    <v-col>
      <v-text-field
        :disabled="props.isRoomIdDisabled"
        label="Raum"
        v-model="roomId"
        required
        :rules="textFieldRules"
      />
      <v-text-field label="Name" v-model="name" required :rules="textFieldRules" />
    </v-col>

    <v-radio-group label="Deine Rolle" v-model="role">
      <v-radio label="Product Owner" :value="Role.ProductOwner"></v-radio>
      <v-radio label="Entwickler" :value="Role.Developer"></v-radio>
    </v-radio-group>

    <v-col v-if="props.errorMessage !== '' && props.errorMessage !== null">
      <v-alert color="error" :text="props.errorMessage" />
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
