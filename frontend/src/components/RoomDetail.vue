<script setup lang="ts">
import {Role, RoundState, type UserOverview} from "@/components/types";
import UserBox from "@/components/UserBox.vue";
import CommandCenter from "@/components/CommandCenter.vue";
import {computed, ref} from "vue";

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
  (e: "reveal"): void;
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
        <v-card :title="`Aktuelles Ticket zum schätzen: ${props.ticketToGuess}`">
          <v-list>
            <v-list-item v-for="developer in props.usersInRoom.developerList" :key="developer.name">
              <v-list-item-title>
                {{ developer.name }}
                <v-icon color="green" v-if="developer.guess !== 0">mdi-check-circle</v-icon>
                <v-icon v-else>mdi-help-circle</v-icon>
              </v-list-item-title>
            </v-list-item>
          </v-list>

          <v-card-actions v-if="roundIsFinished && userIsProductOwner">
            <v-spacer />
            <v-btn color="primary" @click="emit('reveal')">Auflösen</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-15">
      <v-col cols="12">
        <command-center
          :user-role="props.userRole"
          :round-state="props.roundState"
          :guess="props.guess"
          :ticket-to-guess="props.ticketToGuess"
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
