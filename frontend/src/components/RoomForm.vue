<script setup lang="ts">
import { Role } from "@/components/types";
import { computed, ref } from "vue";

type Props = {
  title: string;
  maxAllowedChars?: number;
  errorMessage?: string | null;
  showPasswordInput?: boolean;
};

const name = defineModel("name", { required: true, type: String, default: "" });
const role = defineModel("role", { required: true, type: String, default: Role.Empty });
const password = defineModel("password", { type: String, default: "", required: false });
const formIsValid = ref(false);

const props = withDefaults(defineProps<Props>(), {
  errorMessage: null,
  showPasswordInput: false,
  maxAllowedChars: 15
});
const emits = defineEmits<{
  (e: "submit"): void;
}>();

const textFieldRules = computed(() => [
  (value: string) => !!value || "Can not be empty",
  (value: string) =>
    (value && value.length <= props.maxAllowedChars) ||
    `Only ${props.maxAllowedChars} chars allowed`,
]);
</script>

<template>
  <v-card :title="props.title">
    <v-card-text>
      <slot name="teaser" />
      <v-form
        v-model="formIsValid"
        fast-fail
        validate-on="input"
        @submit.prevent="emits('submit')"
      >
        <v-text-field
          v-model="name"
          label="Name"
          required
          :rules="textFieldRules"
        />

        <v-radio-group
          v-model="role"
          label="Your role"
          :rules="[(value) => !!value || 'Can not be empty']"
        >
          <v-radio
            label="Product Owner"
            :value="Role.ProductOwner"
          />
          <v-radio
            label="Developer"
            :value="Role.Developer"
          />
        </v-radio-group>

        <v-text-field
          v-if="props.showPasswordInput"
          v-model="password"
          type="password"
          label="Password"
          required
          :rules="textFieldRules"
        />

        <v-alert
          v-if="props.errorMessage"
          color="error"
          :text="props.errorMessage"
        />

        <v-btn
          class="float-right"
          type="submit"
          color="primary"
          prepend-icon="mdi-connection"
          :disabled="!formIsValid"
        >
          Connect
        </v-btn>
      </v-form>
    </v-card-text>
  </v-card>
</template>

<style scoped></style>
