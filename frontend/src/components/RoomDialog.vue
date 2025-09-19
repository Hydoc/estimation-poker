<script setup lang="ts">
import { computed, ref } from "vue";
import { Role } from "@/components/types.ts";

type Props = {
  maxAllowedChars: number;
  activatorText: string;
  cardTitle: string;
  errorMessage?: string;
};
const props = defineProps<Props>();
const emits = defineEmits<{
  (e: "submit"): void;
}>();
const role = defineModel("role", { type: String, default: Role.Empty, required: true });
const name = defineModel("name", { type: String, default: "", required: true });

const showDialog = ref(false);
const formIsValid = ref(false);

const textFieldRules = computed(() => [
  (value: string) => !!value || "Can not be empty",
  (value: string) =>
    (value && value.length <= props.maxAllowedChars) ||
    `Only ${props.maxAllowedChars} chars allowed`,
]);
</script>

<template>
  <v-dialog
    v-model="showDialog"
    width="500"
  >
    <template #activator="{ props: activatorProps }">
      <v-btn
        v-bind="activatorProps"
        color="primary"
        :text="props.activatorText"
      />
    </template>

    <v-card :title="props.cardTitle">
      <v-card-text>
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
  </v-dialog>
</template>

<style scoped></style>
