import { isJust, just, type Maybe, nothing } from "@kaumlaut/pure/maybe";
import { computed, type ComputedRef, ref } from "vue";

export type UseWebsocket = {
  isConnected: ComputedRef<boolean>;
  connect(url: string, onMessage: (message: MessageEvent) => Promise<void>): Promise<boolean>;
  disconnect(): void;
  send<T extends Record<string, any>>(message: T): void;
};

export function useWebsocket(): UseWebsocket {
  const connection = ref<Maybe<WebSocket>>(nothing());

  const isConnected = computed(() => isJust(connection.value));
  
  function send<T extends Record<string, any>>(message: T) {
    if (!isJust(connection.value)) {
      throw new Error("Can not send message without a connection");
    }
    connection.value.value.send(JSON.stringify(message));
  }

  async function connect(
    url: string,
    onMessage: (message: MessageEvent) => Promise<void>,
  ): Promise<boolean> {
    let wsUrl = `wss://${url}`;
    if (window.location.protocol !== "https:") {
      wsUrl = `ws://${url}`;
    }

    connection.value = just(new WebSocket(wsUrl));
    const connected = await waitForOpenConnection();
    if (!connected) {
      disconnect();
      return false;
    }

    connection.value.value!.onerror = () => {
      if (isJust(connection.value)) {
        connection.value.value.close();
      }
    };

    connection.value.value!.onclose = () => {
      if (isJust(connection.value)) {
        connection.value.value.close();
      }
    };

    if (isJust(connection.value)) {
      connection.value.value.onmessage = onMessage;
    }

    return true;
  }

  function disconnect() {
    if (isJust(connection.value)) {
      connection.value.value.close();
      connection.value = nothing();
    }
  }

  async function waitForOpenConnection(): Promise<boolean> {
    return new Promise((resolve) => {
      if (!isJust(connection.value)) {
        resolve(false);
        return;
      }
      if (connection.value.value.readyState !== connection.value.value.OPEN) {
        connection.value.value.addEventListener("open", () => {
          resolve(true);
        });
      } else {
        resolve(true);
      }
    });
  }

  return {
    connect,
    disconnect,
    isConnected,
    send,
  };
}
