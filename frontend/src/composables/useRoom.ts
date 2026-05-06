import { computed, type ComputedRef, ref } from "vue";
import { type Maybe, nothing } from "@kaumlaut/pure/maybe";
import type { RoomState } from "@/types/room.ts";

export type UseRoom = {
    roomState: ComputedRef<RoomState>;
    setRoomId: (id: Maybe<string>) => void;
};

export function useRoom(): UseRoom {
    const roomId = ref<Maybe<string>>(nothing());
    
    const roomState = computed(() => ({
        roomId: roomId.value,
    }));
    
    function setRoomId(id: Maybe<string>) {
        roomId.value = id;
    }
    
    // async function fetchRoomState() {
    //     state.value = load();
    //    
    //     const response = await fetch(`/v1/room/${roomId}/state`);
    //    
    //     state.value = attemptErrorAware(isRoomState)(d)
    // }
    
    return {
        roomState,
        setRoomId,
    };
}