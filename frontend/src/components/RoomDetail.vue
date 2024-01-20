<script setup lang="ts">
import { Role, RoundState, type UserOverview } from "@/components/types";
import UserBox from "@/components/UserBox.vue";
import CommandCenter from "@/components/CommandCenter.vue";
import { computed, ref } from "vue";
import RoundOverview from "@/components/RoundOverview.vue";

type Props = {
  roomId: string;
  usersInRoom: UserOverview;
  currentUsername: string;
  userRole: Role;
  roundState: RoundState;
  ticketToGuess: string;
  guess: number;
  showAllGuesses: boolean;
};

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "estimate", ticket: string): void;
  (e: "guess", guess: number): void;
  (e: "reveal"): void;
  (e: "new-round"): void;
}>();
const showSnackbar = ref(false);
const roundIsFinished = computed(() => props.roundState === RoundState.End);
const userIsProductOwner = computed(() => props.userRole === Role.ProductOwner);

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
        <round-overview
          :round-is-finished="roundIsFinished"
          :show-all-guesses="props.showAllGuesses"
          :developer-list="props.usersInRoom.developerList"
          :ticket-to-guess="props.ticketToGuess"
          :user-is-product-owner="userIsProductOwner"
          @reveal="emit('reveal')"
          @new-round="emit('new-round')"
        />
      </v-col>
    </v-row>

    <v-row class="mt-15">
      <v-col cols="12">
        <command-center
          :user-role="props.userRole"
          :round-state="props.roundState"
          :guess="props.guess"
          :ticket-to-guess="props.ticketToGuess"
          :has-developers-in-room="props.usersInRoom.developerList.length > 0"
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
