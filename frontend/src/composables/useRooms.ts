import { computed, type ComputedRef, ref } from "vue";
import {
  attemptErrorAware,
  fail,
  type FetchState,
  isSuccess,
  none,
} from "@kaumlaut/pure/fetch-state";
import { type ActiveRooms, isActiveRooms, isRoomCreated } from "@/types/rooms.ts";

type UseRoomsState = Readonly<{
  availableActiveRooms: FetchState<ActiveRooms>;
}>;

type UseRooms = {
  roomsState: ComputedRef<UseRoomsState>;
  fetchActiveRooms(): Promise<void>;
  createRoom(name: string): Promise<string>;
};

export function useRooms(): UseRooms {
  const availableActiveRooms = ref<FetchState<ActiveRooms>>(none());

  const roomsState = computed(
    (): UseRoomsState => ({
      availableActiveRooms: availableActiveRooms.value,
    }),
  );

  async function fetchActiveRooms() {
    const response = await fetch("/v1/rooms");

    if (!response.ok) {
      availableActiveRooms.value = fail("error fetching active rooms");
    } else {
      const data = await response.json();
      availableActiveRooms.value = attemptErrorAware(isActiveRooms)(data);
    }
  }

  async function createRoom(name: string): Promise<string> {
    const response = await fetch("/v1/room", {
      method: "POST",
      body: JSON.stringify({
        creator: name,
        guesses: {},
      }),
    });

    if (!response.ok) {
      throw new Error("could not create room");
    }

    const result = attemptErrorAware(isRoomCreated)(await response.json());

    if (!isSuccess(result)) {
      throw new Error("could not create room");
    }

    return result.data.id;
  }

  return {
    roomsState,
    fetchActiveRooms,
    createRoom,
  };
}
