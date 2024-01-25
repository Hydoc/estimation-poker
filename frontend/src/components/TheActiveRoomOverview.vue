<script setup lang="ts">
import { ref } from "vue";
import RoomForm from "@/components/RoomForm.vue";

type Props = {
  activeRooms: string[];
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "join", name: string, role: string): void;
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
    <v-dialog v-model="showDialog" width="auto">
      <v-card title="Zu faul den Raum abzutippen? Na gut.">
        <v-card-text>
          {{ roomToJoin }}
          <room-form role="" name="" room-id="" />
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
            <v-btn append-icon="mdi-location-enter" @click="showDialogForRoom(room)">Beitreten</v-btn>
          </td>
        </tr>
      </tbody>
    </v-table>
  </v-sheet>
</template>

<style scoped></style>
