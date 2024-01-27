import { beforeEach, describe, expect, it, vi } from "vitest";
import { useWebsocketStore } from "../../src/stores/websocket";
import { Role, RoundState } from "../../src/components/types";
import { createPinia, setActivePinia } from "pinia";

const websocketSendSpy = vi.fn();
const websocketCloseSpy = vi.fn();
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
  it("should connect", () => {
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
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
    expect(websocketStore.usersInRoom).deep.equal({
      developerList: [],
      productOwnerList: [],
    });
  });

  it("should connect as developer", () => {
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.Developer, "Test");
    expect(websocketUrl).equal("ws://localhost:3000/api/estimation/room/Test/developer?name=ABC");
  });

  it("should send a message", async () => {
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
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
    const usersInRoom = {
      developerList: [{ name: "C", role: Role.Developer, guess: 0 }],
      productOwnerList: [{ name: "ABC", role: Role.ProductOwner, guess: 0 }],
    };
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => usersInRoom,
    });

    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "join" }),
    });

    expect(websocketStore.usersInRoom).deep.equal(usersInRoom);
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/users");
  });

  it("should fetch users in room when leave message appeared", async () => {
    const usersInRoom = {
      developerList: [{ name: "C", role: Role.Developer, guess: 0 }],
      productOwnerList: [{ name: "ABC", role: Role.ProductOwner, guess: 0 }],
    };
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => usersInRoom,
    });

    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "leave" }),
    });

    expect(websocketStore.usersInRoom).deep.equal(usersInRoom);
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/users");
  });

  it("should fetch users in room when developer-guessed message appeared", async () => {
    const usersInRoom = {
      developerList: [{ name: "C", role: Role.Developer, guess: 0 }],
      productOwnerList: [{ name: "ABC", role: Role.ProductOwner, guess: 0 }],
    };
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => usersInRoom,
    });

    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "developer-guessed" }),
    });

    expect(websocketStore.usersInRoom).deep.equal(usersInRoom);
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/users");
  });

  it("should update round state and ticket to guess when estimate appears", async () => {
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "estimate", data: "WR-1" }),
    });

    expect(websocketStore.ticketToGuess).equal("WR-1");
    expect(websocketStore.roundState).equal(RoundState.InProgress);
  });

  it("should update guess when you-guessed message appeared", async () => {
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "you-guessed", data: 1 }),
    });

    expect(websocketStore.guess).equal(1);
  });

  it("should fetch users in room and end round when everyone-guessed message appeared", async () => {
    const usersInRoom = {
      developerList: [{ name: "C", role: Role.Developer, guess: 0 }],
      productOwnerList: [{ name: "ABC", role: Role.ProductOwner, guess: 0 }],
    };
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => usersInRoom,
    });
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "everyone-guessed" }),
    });
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/users");
    expect(websocketStore.roundState).equal(RoundState.End);
  });

  it("should set show all guesses to true when reveal message appeared", async () => {
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "reveal" }),
    });
    expect(websocketStore.showAllGuesses).to.be.true;
  });

  it("should reset round and fetch users in room when reset-round message appeared", async () => {
    const usersInRoom = {
      developerList: [{ name: "C", role: Role.Developer, guess: 0 }],
      productOwnerList: [{ name: "ABC", role: Role.ProductOwner, guess: 0 }],
    };
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => usersInRoom,
    });
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    websocketStore.ticketToGuess = "BLA-1";
    websocketStore.guess = 9;
    websocketStore.roundState = RoundState.End;
    websocketStore.showAllGuesses = true;
    await websocketOnMessage({
      data: JSON.stringify({ type: "reset-round" }),
    });

    expect(websocketStore.ticketToGuess).equal("");
    expect(websocketStore.guess).equal(0);
    expect(websocketStore.roundState).equal(RoundState.Waiting);
    expect(websocketStore.showAllGuesses).to.be.false;
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/Test/users");
  });

  it("should close when error occured", () => {
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    websocketOnError();
    expect(websocketCloseSpy).toHaveBeenCalledOnce();
  });

  it("should close when server closes the connection", () => {
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    websocketOnClose();
    expect(websocketCloseSpy).toHaveBeenCalledOnce();
  });

  it("should close when calling disconnect", () => {
    const websocketStore = useWebsocketStore();
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    websocketStore.disconnect();
    expect(websocketCloseSpy).toHaveBeenCalledOnce();
  });

  it("should reset usersInRoom when response is not ok", async () => {
    const usersInRoom = {
      developerList: [{ name: "C", role: Role.Developer, guess: 0 }],
      productOwnerList: [{ name: "ABC", role: Role.ProductOwner, guess: 0 }],
    };
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
    });
    const websocketStore = useWebsocketStore();
    websocketStore.usersInRoom = usersInRoom;
    websocketStore.connect("ABC", Role.ProductOwner, "Test");
    await websocketOnMessage({
      data: JSON.stringify({ type: "leave" }),
    });

    expect(websocketStore.usersInRoom).deep.equal({
      developerList: [],
      productOwnerList: [],
    });
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
      json: () => ["Test room"],
    });
    const websocketStore = useWebsocketStore();
    const actual = await websocketStore.fetchActiveRooms();
    expect(actual).deep.equal(["Test room"]);
    expect(global.fetch).toHaveBeenNthCalledWith(1, "/api/estimation/room/rooms");
  });
});
