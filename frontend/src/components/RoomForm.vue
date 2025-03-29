<script setup lang="ts">
import { Role } from "@/components/types";
import { computed, ref } from "vue";

type Props = {
  errorMessage?: string | null;
  isRoomIdDisabled?: boolean;
};

const maxAllowedChars = 25;
const name = defineModel("name", { required: true, type: String, default: "" });
const roomId = defineModel("roomId", { required: true, type: String, default: "" });
const role = defineModel("role", { required: true, type: String, default: Role.Empty });
const formIsValid = ref(false);

const props = withDefaults(defineProps<Props>(), {
  errorMessage: null,
  isRoomIdDisabled: false,
});
const emit = defineEmits<{
  (e: "submit"): void;
}>();

const textFieldRules = computed(() => [
  (value: string) => !!value || "Fehler: Hier müsste eigentlich was stehen",
  (value: string) => (value && value.length <= maxAllowedChars) || `Fehler: Maximallänge von ${maxAllowedChars} darf nicht überschritten werden`
]);
</script>

<template>
  <v-form
    v-model="formIsValid"
    :fast-fail="true"
    validate-on="input"
    @submit.prevent="emit('submit')"
  >
    <v-col>
      <v-text-field
        v-model="roomId"
        :disabled="props.isRoomIdDisabled"
        label="Raum"
        required
        :rules="textFieldRules"
      />
      <v-text-field
        v-model="name"
        label="Name"
        required
        :rules="textFieldRules"
      />
    </v-col>

    <v-radio-group
      v-model="role"
      label="Deine Rolle"
      :rules="[(value) => !!value || 'Fehler: Hier müsste eigentlich was stehen']"
    >
      <v-radio
        label="Product Owner"
        :value="Role.ProductOwner"
      />
      <v-radio
        label="Entwickler"
        :value="Role.Developer"
      />
    </v-radio-group>

    <v-col v-if="props.errorMessage !== '' && props.errorMessage !== null">
      <v-alert
        color="error"
        :text="props.errorMessage"
      />
    </v-col>

    <v-col class="text-right">
      <v-btn
        type="submit"
        color="primary"
        prepend-icon="mdi-connection"
        :disabled="!formIsValid"
      >
        Verbinden
      </v-btn>
    </v-col>
  </v-form>
</template>

<style scoped></style>
