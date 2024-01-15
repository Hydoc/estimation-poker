import { computed, ref } from "vue";
import type { Ref } from "vue";
import { defineStore } from "pinia";
import { Role } from "@/components/types";
import type { Developer, ProductOwner } from "@/components/types";

type WebsocketStore = {
  connect(name: string, role: string, roomId: string): void;
  username: Ref<string>;
  isConnected: Ref<boolean>;
  usersInRoom: Ref<(Developer | ProductOwner)[]>;
  roomId: Ref<string>;
};

enum WebsocketMessage {
  Join = "join",
  Leave = "leave",
}

export const useWebsocketStore = defineStore("websocket", (): WebsocketStore => {
  const username = ref("");
  const userRole: Ref<Role> = ref(Role.Empty);
  const userRoomId = ref("");
  const websocket: Ref<WebSocket | null> = ref(null);
  const usersInRoom = ref([]);

  const isConnected = computed(() => websocket.value !== null);

  function connect(name: string, role: Role, roomId: string): void {
    username.value = name;
    userRole.value = role;
    userRoomId.value = roomId;

    const roleUrl = role === Role.Developer ? "developer" : "product-owner";
    websocket.value = new WebSocket(`ws://localhost:8090/room/${roomId}/${roleUrl}?name=${name}`);

    websocket.value!.onerror = () => {
      websocket.value!.close();
      setTimeout(() => connect(name, role, roomId), 3000);
    }

    websocket.value!.onmessage = async (message: MessageEvent) => {
      switch (message.data) {
        case WebsocketMessage.Leave:
        case WebsocketMessage.Join:
          await fetchUsersInRoom();
          break;
      }
      console.log("Message flew", message);
    }
  }

  async function fetchUsersInRoom() {
    const response = await fetch(`http://localhost:8090/room/${userRoomId.value}/users`);
    if (!response.ok) {
      usersInRoom.value = [];
      return;
    }

    usersInRoom.value = await response.json();
  }

  return { connect, isConnected, usersInRoom, roomId: userRoomId, username };
});
