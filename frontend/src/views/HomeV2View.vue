<script setup lang="ts">
import {ref} from "vue";

const displayName = ref("");

const rules = [
  (value: string) => {
    if (value) return true
    return "Is required"
  },
]

function connect() {}
function create() {}
</script>

<template>
  <main>
    <v-card title="To which room would you like to connect?">
      <v-card-subtitle>If there is already a room you can directly connect through the link.</v-card-subtitle>
      <v-card-text>
        <v-form @submit:prevent="connect">
          <v-text-field
            placeholder="Room id"
            :rules="rules"
            required
          />
          <v-text-field
            v-model.trim="displayName"
            placeholder="Your display name"
            :rules="rules"
            required
          />
          
          <v-card-actions>
            <v-spacer />

            <v-btn
              color="primary"
              prepend-icon="mdi-connection"
            >
              Connect
            </v-btn>
          </v-card-actions>
        </v-form>
        
        <v-spacer />
        
        <div>
          <p class="text-title-medium">
            As an alternative you can also create a new one
          </p>
          <v-form @submit:prevent="create">
            <v-text-field
              v-model.trim="displayName"
              :rules="rules"
              placeholder="Your display name"
            />
            
            <v-card-actions>
              <v-spacer />
              <v-btn
                color="primary"
                prepend-icon="mdi-plus"
              >
                Create a new room
              </v-btn>
            </v-card-actions>
          </v-form>
        </div>
      </v-card-text>
    </v-card>
  </main>
</template>

<style scoped>

</style>