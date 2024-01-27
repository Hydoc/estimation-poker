<script setup lang="ts">
import { type Ref, ref } from "vue";
import RoomForm from "@/components/RoomForm.vue";
import { Role } from "@/components/types";

type Props = {
  activeRooms: string[];
  errorMessage?: string | null;
};

const name = ref("");
const role: Ref<Role> = ref(Role.Empty);

const props = withDefaults(defineProps<Props>(), {
  errorMessage: null,
});
const emit = defineEmits<{
  (e: "join", roomId: string, name: string, role: Role): void;
}>();

const roomToJoin = ref("");
const showDialog = ref(false);

function showDialogForRoom(room: string) {
  roomToJoin.value = room;
  showDialog.value = true;
}
</script>

<template>
  <v-sheet>
    <v-dialog v-model="showDialog" width="500">
      <v-card title="Raum beitreten">
        <v-card-text>
          <room-form
            :is-room-id-disabled="true"
            v-model:role="role"
            v-model:name="name"
            :room-id="roomToJoin"
            :error-message="props.errorMessage"
            @submit="emit('join', roomToJoin, name, role)"
          />
        </v-card-text>
      </v-card>
    </v-dialog>

    <h2>Bereits erstellte Räume</h2>
    <v-table>
      <thead>
        <tr>
          <th>Raum</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="room in props.activeRooms" :key="room">
          <td>{{ room }}</td>
          <td class="text-right">
            <v-btn append-icon="mdi-location-enter" @click="showDialogForRoom(room)"
              >Beitreten</v-btn
            >
          </td>
        </tr>
      </tbody>
    </v-table>
  </v-sheet>
</template>

<style scoped></style>
