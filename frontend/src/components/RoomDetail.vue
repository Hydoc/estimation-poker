<script setup lang="ts">
import { Role, RoundState, type UserOverview } from "@/components/types";
import UserBox from "@/components/UserBox.vue";
import CommandCenter from "@/components/CommandCenter.vue";
import { ref } from "vue";

type Props = {
  roomId: string;
  usersInRoom: UserOverview;
  currentUsername: string;
  userRole: Role;
  roundState: RoundState;
  ticketToGuess: string;
  guess: number;
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "estimate", ticket: string): void;
  (e: "guess", guess: number): void;
}>();
const showSnackbar = ref(false);

function copyRoomName() {
  showSnackbar.value = true;
  navigator.clipboard.writeText(props.roomId);
}
</script>

<template>
  <h1>
    Raum: {{ props.roomId }}
    <v-icon title="Raum kopieren" size="x-small" @click="copyRoomName">mdi-content-copy</v-icon>
  </h1>

  <v-container width="200">
    <v-row>
      <v-col>
        <user-box
          title="Product Owner"
          :user-list="usersInRoom.productOwnerList"
          :current-username="currentUsername"
        />
      </v-col>
      <v-col>
        <user-box
          title="Entwickler"
          :user-list="usersInRoom.developerList"
          :current-username="currentUsername"
        />
      </v-col>
    </v-row>

    <v-row class="mt-15" v-if="ticketToGuess !== ''">
      <v-col cols="12">
        <p>Aktuelles Ticket zum schätzen: {{ props.ticketToGuess }}</p>
      </v-col>
    </v-row>

    <v-row class="mt-15">
      <v-col cols="12">
        <command-center
          :user-role="props.userRole"
          :round-state="props.roundState"
          :guess="props.guess"
          @estimate="emit('estimate', $event)"
          @guess="emit('guess', $event)"
        />
      </v-col>
    </v-row>
  </v-container>
  <v-snackbar :timeout="3000" v-model="showSnackbar"
    >Raum in die Zwischenablage kopiert!</v-snackbar
  >
</template>

<style scoped></style>
