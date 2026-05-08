import { defineStore } from "pinia";
import { useRoom } from "@/composables/useRoom.ts";
import { useRooms } from "@/composables/useRooms.ts";

export const useEstimationStore = defineStore("estimation", () => {
  const room = useRoom();
  const rooms = useRooms();

  return {
    ...room,
    ...rooms,
  };
});
