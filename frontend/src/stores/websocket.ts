import type { Ref } from "vue";
import { computed, ref } from "vue";
import { defineStore } from "pinia";
import type { UserOverview } from "@/components/types";
import { Role } from "@/components/types";

type WebsocketStore = {
  connect(name: string, role: string, roomId: string): void;
  disconnect(): void;
  userExistsInRoom(name: string, roomId: string): Promise<boolean>;
  username: Ref<string>;
  isConnected: Ref<boolean>;
  usersInRoom: Ref<UserOverview>;
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
  const usersInRoom: Ref<UserOverview> = ref({
    developerList: [],
    productOwnerList: [],
  });

  const isConnected = computed(() => websocket.value !== null);

  function disconnect() {
    if (websocket.value) {
      websocket.value!.close();
      websocket.value = null;
    }
  }

  function connect(name: string, role: Role, roomId: string): void {
    username.value = name;
    userRole.value = role;
    userRoomId.value = roomId;

    const roleUrl = role === Role.Developer ? "developer" : "product-owner";
    websocket.value = new WebSocket(`ws://localhost:8090/room/${roomId}/${roleUrl}?name=${name}`);

    websocket.value!.onerror = () => {
      websocket.value!.close();
    };

    websocket.value!.onclose = () => {
      websocket.value!.close();
    };

    websocket.value!.onmessage = async (message: MessageEvent) => {
      switch (message.data) {
        case WebsocketMessage.Leave:
        case WebsocketMessage.Join:
          await fetchUsersInRoom();
          break;
      }
      console.log("Message flew", message);
    };
  }

  async function userExistsInRoom(name: string, roomId: string): Promise<boolean> {
    const response = await fetch(`http://localhost:8090/room/${roomId}/users/exists?name=${name}`);
    return (await response.json()).exists;
  }

  async function fetchUsersInRoom() {
    const response = await fetch(`http://localhost:8090/room/${userRoomId.value}/users`);
    if (!response.ok) {
      usersInRoom.value = {
        productOwnerList: [],
        developerList: [],
      };
      return;
    }

    usersInRoom.value = await response.json();
  }

  return {
    connect,
    disconnect,
    isConnected,
    usersInRoom,
    roomId: userRoomId,
    username,
    userExistsInRoom,
  };
});
