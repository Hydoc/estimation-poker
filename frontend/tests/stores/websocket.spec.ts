import { beforeEach, describe, expect, it, vi } from "vitest";
import { useWebsocketStore } from "../../src/stores/websocket";
import { Role, RoundState } from "../../src/components/types";
import { createPinia, setActivePinia } from "pinia";

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

beforeEach(() => {
  setActivePinia(createPinia());
  websocketSendSpy.mockReset();
  websocketCloseSpy.mockReset();
  websocketUrl = "";
});
describe("Websocket Store", () => {
  it("should connect", async () => {
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    expect(websocketUrl).equal(
      "ws://localhost:3000/api/estimation/room/Test/product-owner?name=ABC",
    );
    expect(websocketStore.username).equal("ABC");
    expect(websocketStore.userRole).equal(Role.ProductOwner);
    expect(websocketStore.roomId).equal("Test");
    expect(websocketStore.showAllGuesses).to.be.false;
    expect(websocketStore.isConnected).to.be.true;
    expect(websocketStore.ticketToGuess).equal("");
    expect(websocketStore.guess).equal(0);
    expect(websocketStore.didSkip).false;
    expect(websocketStore.usersInRoom).deep.equal([]);
  });

  it("should connect as developer", async () => {
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.Developer, "Test");
    expect(websocketUrl).equal("ws://localhost:3000/api/estimation/room/Test/developer?name=ABC");
  });

  it("should send a message", async () => {
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    websocketStore.send({
      type: "estimate",
      data: "WR-123",
    });
    expect(websocketSendSpy).toHaveBeenNthCalledWith(
      1,
      JSON.stringify({
        type: "estimate",
        data: "WR-123",
      }),
    );
  });

  it("should throw an error when trying to send a message but connection is not established", () => {
    const websocketStore = useWebsocketStore();
    expect(() => {
      websocketStore.send({
        type: "estimate",
        data: "WR-123",
      });
    }).toThrow(new Error("Can not send message without a connection"));
  });

  it("should fetch users in room when join message appeared", async () => {
    const usersInRoom = [
      { name: "C", role: Role.Developer, guess: 0 },
      { name: "ABC", role: Role.ProductOwner, guess: 0 },
    ];
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => usersInRoom,
    });

    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "join" }),
    });

    expect(websocketStore.usersInRoom).deep.equal(usersInRoom);
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/users");
  });

  it("should fetch users in room when leave message appeared", async () => {
    const usersInRoom = [
      { name: "C", role: Role.Developer, guess: 0 },
      { name: "ABC", role: Role.ProductOwner, guess: 0 },
    ];
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => usersInRoom,
    });

    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "leave" }),
    });

    expect(websocketStore.usersInRoom).deep.equal(usersInRoom);
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/users");
  });

  it("should fetch users in room when developer-guessed message appeared", async () => {
    const usersInRoom = [
      { name: "C", role: Role.Developer, guess: 0 },
      { name: "ABC", role: Role.ProductOwner, guess: 0 },
    ];
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => usersInRoom,
    });

    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "developer-guessed" }),
    });

    expect(websocketStore.usersInRoom).deep.equal(usersInRoom);
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/users");
  });

  it("should update round state and ticket to guess when estimate appears", async () => {
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "estimate", data: "WR-1" }),
    });

    expect(websocketStore.ticketToGuess).equal("WR-1");
    expect(websocketStore.roundState).equal(RoundState.InProgress);
  });

  it("should update guess when you-guessed message appeared", async () => {
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "you-guessed", data: 1 }),
    });

    expect(websocketStore.guess).equal(1);
  });

  it("should update didSkip when you-skipped message appeared", async () => {
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "you-skipped" }),
    });

    expect(websocketStore.didSkip).true;
  });

  it("should fetch users in room and end round when everyone-guessed message appeared", async () => {
    const usersInRoom = [
      { name: "C", role: Role.Developer, guess: 0 },
      { name: "ABC", role: Role.ProductOwner, guess: 0 },
    ];
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => usersInRoom,
    });
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "everyone-done" }),
    });
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/users");
    expect(websocketStore.roundState).equal(RoundState.End);
  });

  it("should set show all guesses to true when reveal message appeared", async () => {
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "reveal", data: [] }),
    });
    expect(websocketStore.showAllGuesses).to.be.true;
  });

  it("should fetch room is locked when room-locked message appeared", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => ({ isLocked: true }),
    });
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "room-locked" }),
    });

    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/state");
    expect(websocketStore.roomIsLocked).to.be.true;
  });

  it("should fetch room is locked when room-opened message appeared", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => ({ isLocked: false }),
    });
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "room-opened" }),
    });

    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/state");
    expect(websocketStore.roomIsLocked).to.be.false;
  });

  it("should reset round and fetch users in room when new-round message appeared", async () => {
    const usersInRoom = [
      { name: "C", role: Role.Developer, guess: 0 },
      { name: "ABC", role: Role.ProductOwner, guess: 0 },
    ];
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => usersInRoom,
    });
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    websocketStore.ticketToGuess = "BLA-1";
    websocketStore.guess = 9;
    websocketStore.roundState = RoundState.End;
    websocketStore.showAllGuesses = true;
    await websocketOnMessage({
      data: JSON.stringify({ type: "new-round" }),
    });

    expect(websocketStore.ticketToGuess).equal("");
    expect(websocketStore.guess).equal(0);
    expect(websocketStore.roundState).equal(RoundState.Waiting);
    expect(websocketStore.showAllGuesses).to.be.false;
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/users");
  });

  it("should close when error occured", async () => {
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    websocketOnError();
    expect(websocketCloseSpy).toHaveBeenCalledOnce();
  });

  it("should close when server closes the connection", async () => {
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    websocketOnClose();
    expect(websocketCloseSpy).toHaveBeenCalledOnce();
  });

  it("should close when calling disconnect", async () => {
    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    websocketStore.disconnect();
    expect(websocketCloseSpy).toHaveBeenCalledOnce();
  });

  it("should reset usersInRoom when response is not ok", async () => {
    const usersInRoom = [
      { name: "C", role: Role.Developer, guess: 0 },
      { name: "ABC", role: Role.ProductOwner, guess: 0 },
    ];
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
    });
    const websocketStore = useWebsocketStore();
    websocketStore.usersInRoom = usersInRoom;
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "leave" }),
    });

    expect(websocketStore.usersInRoom).deep.equal([]);
  });

  it("should return true when user in room exists", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      json: () => ({ exists: true }),
    });
    const websocketStore = useWebsocketStore();
    const actual = await websocketStore.userExistsInRoom("Bla", "Blub");
    expect(actual).to.be.true;
    expect(global.fetch).toHaveBeenNthCalledWith(
      1,
      "/api/estimation/room/Blub/users/exists?name=Bla",
    );
  });

  it("should return false when user in room does not exist", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      json: () => ({ exists: false }),
    });
    const websocketStore = useWebsocketStore();
    const actual = await websocketStore.userExistsInRoom("Bla", "Blub");
    expect(actual).to.be.false;
    expect(global.fetch).toHaveBeenNthCalledWith(
      1,
      "/api/estimation/room/Blub/users/exists?name=Bla",
    );
  });

  it("should return false when round in room not in progress", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      json: () => ({ inProgress: false }),
    });
    const websocketStore = useWebsocketStore();
    const actual = await websocketStore.isRoundInRoomInProgress("Blub");
    expect(actual).to.be.false;
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Blub/state");
  });

  it("should return true when round in room in progress", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      json: () => ({ inProgress: true }),
    });
    const websocketStore = useWebsocketStore();
    const actual = await websocketStore.isRoundInRoomInProgress("Blub");
    expect(actual).to.be.true;
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Blub/state");
  });

  it("should fetch active rooms", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => ({
        rooms: [
          {
            id: "any-id",
            playerCount: 1,
          },
        ],
      }),
    });
    const websocketStore = useWebsocketStore();
    const actual = await websocketStore.fetchActiveRooms();
    expect(actual).deep.equal([
      {
        id: "any-id",
        playerCount: 1,
      },
    ]);
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/rooms");
  });

  it("should fetch passwordMatchesRoom when password matches", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => ({ ok: true }),
    });
    const websocketStore = useWebsocketStore();
    const matches = await websocketStore.passwordMatchesRoom("abc", "top secret");
    expect(matches).to.be.true;
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/abc/authenticate", {
      method: "POST",
      body: JSON.stringify({ password: "top secret" }),
    });
  });

  it("should fetch passwordMatchesRoom when password does not match", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => ({ ok: false }),
    });
    const websocketStore = useWebsocketStore();
    const matches = await websocketStore.passwordMatchesRoom("abc", "top secret");
    expect(matches).to.be.false;
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/abc/authenticate", {
      method: "POST",
      body: JSON.stringify({ password: "top secret" }),
    });
  });

  it("should fetch passwordMatchesRoom when error occurred", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
    });
    const websocketStore = useWebsocketStore();
    const matches = await websocketStore.passwordMatchesRoom("abc", "top secret");
    expect(matches).to.be.false;
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/abc/authenticate", {
      method: "POST",
      body: JSON.stringify({ password: "top secret" }),
    });
  });

  it("should fetch permissions", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => ({
        permissions: {
          room: {
            canLock: true,
            key: "abc",
          },
        },
      }),
    });

    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketStore.fetchPermissions();

    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/ABC/permissions");
    expect(websocketStore.permissions).deep.equal({
      room: {
        canLock: true,
        key: "abc",
      },
    });
  });

  it("should fetch permissions when response not ok", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
    });

    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketStore.fetchPermissions();

    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/ABC/permissions");
    expect(websocketStore.permissions).deep.equal({
      room: {
        canLock: false,
      },
    });
  });

  it("should fetch roomIsLocked", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => ({
        isLocked: true,
      }),
    });

    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    const actual = await websocketStore.fetchRoomIsLocked("Test");
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/state");
    expect(actual).to.be.true;
  });

  it("should fetch roomIsLocked when response not ok", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
    });

    const websocketStore = useWebsocketStore();
    await websocketStore.connect("ABC", Role.ProductOwner, "Test");
    const actual = await websocketStore.fetchRoomIsLocked("Test");
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/state");
    expect(actual).to.be.false;
  });

  it("should fetch possible guesses", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => [
        { guess: 1, description: "Bis zu 4 Std." },
        { guess: 2, description: "Bis zu 8 Std." },
        { guess: 3, description: "Bis zu 3 Tagen" },
        { guess: 4, description: "Bis zu 5 Tagen" },
        { guess: 5, description: "Mehr als 5 Tage" },
      ],
    });
    const websocketStore = useWebsocketStore();
    await websocketStore.fetchPossibleGuesses();
    expect(websocketStore.possibleGuesses).deep.equal([
      { guess: 1, description: "Bis zu 4 Std." },
      { guess: 2, description: "Bis zu 8 Std." },
      { guess: 3, description: "Bis zu 3 Tagen" },
      { guess: 4, description: "Bis zu 5 Tagen" },
      { guess: 5, description: "Mehr als 5 Tage" },
    ]);
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/possible-guesses");
  });

  it("should reset possible guesses when error occurred", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
    });
    const websocketStore = useWebsocketStore();
    await websocketStore.fetchPossibleGuesses();
    expect(websocketStore.possibleGuesses).deep.equal([]);
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/possible-guesses");
  });
});
