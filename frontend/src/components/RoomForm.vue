<script setup lang="ts">
import { Role } from "@/components/types";
import { computed, ref } from "vue";

type Props = {
  errorMessage?: string | null;
  isRoomIdDisabled?: boolean;
};

const maxAllowedChars = 40;
const form = ref();
const name = defineModel("name", { required: true, type: String, default: "" });
const roomId = defineModel("roomId", { required: true, type: String, default: "" });
const role = defineModel("role", { required: true, type: String, default: Role.Empty });

const props = withDefaults(defineProps<Props>(), {
  errorMessage: null,
  isRoomIdDisabled: false,
});
const emit = defineEmits<{
  (e: "submit"): void;
}>();

const isButtonEnabled = computed(
  () => name.value !== "" && name.value.length <= maxAllowedChars && roomId.value !== "" && roomId.value.length <= maxAllowedChars && role.value !== Role.Empty
);

const textFieldRules = computed(() => [
  (value: string) => !!value || "Fehler: Hier müsste eigentlich was stehen",
  (value: string) => (value && value.length <= maxAllowedChars) || "Fehler: Maximallänge von 40 darf nicht überschritten werden"
]);
</script>

<template>
  <v-form
    ref="form"
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
        :disabled="!isButtonEnabled"
      >
        Verbinden
      </v-btn>
    </v-col>
  </v-form>
</template>

<style scoped></style>
