import { describe, it, expect, vi, beforeAll, beforeEach } from "vitest";
import { useRoom } from "../../src/composables/useRoom";
import { just, nothing } from "@kaumlaut/pure/maybe";
import { Role, RoundState } from "../../src/types/room";
import { none } from "@kaumlaut/pure/fetch-state";

const websocketSendSpy = vi.fn();
const websocketCloseSpy = vi.fn();
const addEventListenerSpy = vi.fn();
let websocketUrl;
let websocketOnMessage;
let websocketOnError;
let websocketOnClose;
// @ts-ignore
global.WebSocket = class {
  constructor(url) {
    websocketUrl = url;
  }
  close = websocketCloseSpy;
  send = websocketSendSpy;
  addEventListener = addEventListenerSpy;
  set onmessage(handler) {
    websocketOnMessage = handler;
  }

  set onerror(handler) {
    websocketOnError = handler;
  }

  set onclose(handler) {
    websocketOnClose = handler;
  }
};

describe("useRoom", () => {
  describe("state", () => {
    it("should have default state", () => {
      const composable = useRoom();

      expect(composable.roomState.value).deep.equal({
        id: nothing(),
        guess: nothing(),
        role: nothing(),
        name: nothing(),
        doSkip: false,
        issueToGuess: nothing(),
        roundState: RoundState.Waiting,
        users: none(),
        showAllGuesses: false,
        roomIsLocked: false,
        roundInProgress: false,
        developerDone: [],
        issues: [],
        isConnected: false,
        permissions: { room: { canLock: false } },
        possibleGuesses: [],
      });
    });
  });

  describe("joinRoom", () => {
    it("should join", async () => {
      const composable = useRoom();
      const name = "Tester";
      const role = Role.Developer;
      const roomId = "an-id";

      await composable.joinRoom(name, role, roomId);

      expect(composable.roomState.value.id).deep.equal(just(roomId));
      expect(composable.roomState.value.role).deep.equal(just(role));
      expect(composable.roomState.value.name).deep.equal(just(name));
    });
  });

  describe("send", () => {
    it("should send", async () => {
      const composable = useRoom();
      const name = "Tester";
      const role = Role.Developer;
      const roomId = "an-id";
      await composable.joinRoom(name, role, roomId);

      const message = {
        type: "skip",
        data: "",
      };

      composable.send(message);

      expect(websocketSendSpy).toHaveBeenNthCalledWith(1, JSON.stringify(message));
    });

    it("should throw error when trying to send a message while not connected", async () => {
      try {
        useRoom().send({});
      } catch (e) {
        expect(e.message).equal("Can not send message without a connection");
      }
    });
  });

  // describe("roomMetadata", () => {
  //     it("should fetch", async () => {
  //         // @ts-ignore
  //         global.fetch = vi.fn();
  //
  //
  //     });
  //
  //     it("should throw error if response is not ok", async () => {
  //
  //     });
  //
  //     it("throw error when response is of wrong type", async () => {
  //
  //     });
  // });
});
