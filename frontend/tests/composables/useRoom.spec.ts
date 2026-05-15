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

  describe("roomMetadata", () => {
    it("should fetch", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: true,
        json: () =>
          Promise.resolve({
            exists: true,
            isLocked: false,
          }),
      }));

      const composable = useRoom();
      const roomId = "my-id";

      const metadata = await composable.roomMetadata(roomId);

      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/room/my-id/metadata");
      expect(metadata.exists).to.be.true;
      expect(metadata.isLocked).to.be.false;
    });

    it("should throw error if response is not ok", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: false,
      }));

      const composable = useRoom();
      const roomId = "my-id";

      await expect(composable.roomMetadata(roomId)).rejects.toThrow(
        "Could not fetch room metadata",
      );
      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/room/my-id/metadata");
    });

    it("throw error when response is of wrong type", async () => {
      console.error = vi.fn();
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: true,
        json: () =>
          Promise.resolve({
            invalid: "OK",
          }),
      }));

      const composable = useRoom();
      const roomId = "my-id";

      await expect(composable.roomMetadata(roomId)).rejects.toThrow("Room metadata is invalid");
      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/room/my-id/metadata");
      expect(console.error).toHaveBeenNthCalledWith(1, [
        "Object does not have key exists",
        "Object does not have key isLocked",
      ]);
    });
  });

  describe("connectionState", () => {
    it("should fetch", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: true,
        json: () =>
          Promise.resolve({
            canConnect: false,
            reason: "wrong password",
          }),
      }));

      const composable = useRoom();
      const roomId = "my-id";
      const username = "Tester";
      const password = "";

      const connectionState = await composable.connectionState(roomId, username, password);

      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/room/my-id/connection-state", {
        method: "POST",
        body: JSON.stringify({ username, password }),
      });
      expect(connectionState.canConnect).to.be.false;
      expect(connectionState.reason).equal("wrong password");
    });

    it("should throw error if response is not ok", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: false,
        json: () =>
          Promise.resolve({
            canConnect: false,
            reason: "wrong password",
          }),
      }));

      const composable = useRoom();
      const roomId = "my-id";
      const username = "Tester";
      const password = "";

      await expect(composable.connectionState(roomId, username, password)).rejects.toThrowError(
        "Could not fetch connection state",
      );
      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/room/my-id/connection-state", {
        method: "POST",
        body: JSON.stringify({ username, password }),
      });
    });

    it("throw error when response is of wrong type", async () => {
      console.error = vi.fn();
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: true,
        json: () =>
          Promise.resolve({
            wrong: "type",
          }),
      }));

      const composable = useRoom();
      const roomId = "my-id";
      const username = "Tester";
      const password = "";

      await expect(composable.connectionState(roomId, username, password)).rejects.toThrowError(
        "Connection state is invalid",
      );
      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/room/my-id/connection-state", {
        method: "POST",
        body: JSON.stringify({ username, password }),
      });
      expect(console.error).toHaveBeenNthCalledWith(1, [
        "Object does not have key canConnect",
        "Object does not have key reason",
      ]);
    });
  });

  describe("fetchRoomState", () => {
    it("should fetch and set state", async () => {
      const issues = [
        { title: "Good issue", guess: -1 },
        { title: "Good issue #2", guess: -1 },
      ];
      const possibleGuesses = [
        { guess: 1, description: "4h" },
        { guess: 2, description: "5h" },
      ];
      // @ts-ignore
      global.fetch = vi.fn(() => {
        return {
          ok: true,
          json: () =>
            Promise.resolve({
              issues,
              isLocked: false,
              inProgress: false,
              possibleGuesses,
            }),
        };
      });
      const composable = useRoom();

      await composable.joinRoom("Tester", Role.Developer, "my-id");

      await composable.fetchRoomState();

      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/room/my-id/state");
      expect(composable.roomState.value.issues).deep.equal(issues);
      expect(composable.roomState.value.roomIsLocked).to.be.false;
      expect(composable.roomState.value.roundInProgress).to.be.false;
      expect(composable.roomState.value.possibleGuesses).deep.equal(possibleGuesses);
    });

    it("should throw error when trying to fetch without room id", async () => {
      await expect(useRoom().fetchRoomState()).rejects.toThrowError("Could not fetch room state");
    });

    it("should throw error when response is not ok", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: false,
      }));

      const composable = useRoom();

      await composable.joinRoom("Tester", Role.Developer, "my-id");

      await expect(composable.fetchRoomState()).rejects.toThrowError("Could not fetch room state");
      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/room/my-id/state");
    });

    it("throw error when response is of wrong type", async () => {
      console.error = vi.fn();
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: true,
        json: () =>
          Promise.resolve({
            invalid: "type",
          }),
      }));

      const composable = useRoom();

      await composable.joinRoom("Tester", Role.Developer, "my-id");

      await expect(composable.fetchRoomState()).rejects.toThrowError("Room state is invalid");
      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/room/my-id/state");
      expect(console.error).toHaveBeenNthCalledWith(1, [
        "Object does not have key isLocked",
        "Object does not have key inProgress",
        "Object does not have key issues",
        "Object does not have key possibleGuesses",
      ]);
    });
  });
});
