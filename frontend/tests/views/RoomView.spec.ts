import { beforeEach, describe, expect, it, Mock, vi } from "vitest";
import RoomView from "../../src/views/RoomView.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { useWebsocketStore } from "../../src/stores/websocket";
import { useRoute, useRouter } from "vue-router";
import RoomDetail from "../../src/components/RoomDetail.vue";
import { Role, RoundState } from "../../src/components/types";
import { VBtn, VDialog, VIcon, VTextField, VToolbar } from "vuetify/components";
import { nextTick } from "vue";
import { vuetifyMount } from "../vuetifyMount";
import RoomForm from "../../src/components/RoomForm.vue";

vi.mock("vue-router");

let pinia: TestingPinia;
let websocketStore: ReturnType<typeof useWebsocketStore>;

const ResizeObserverMock = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));
vi.stubGlobal("ResizeObserver", ResizeObserverMock);
vi.stubGlobal("visualViewport", new EventTarget());
beforeEach(() => {
  (useRouter as Mock).mockReturnValue({
    push: vi.fn(),
  });

  (useRoute as Mock).mockReturnValue({
    params: {},
  });

  pinia = createTestingPinia();
  websocketStore = useWebsocketStore(pinia);
  websocketStore.isConnected = true;
  websocketStore.roomId = "ABC";
  websocketStore.username = "test dev";
  websocketStore.userRole = "developer";
  websocketStore.roundState = RoundState.Waiting;
  websocketStore.ticketToGuess = "";
  websocketStore.guess = 0;
  websocketStore.showAllGuesses = false;
  websocketStore.roomExists = vi.fn(() => Promise.resolve(true));
  websocketStore.permissions = {
    room: {
      canLock: true,
      key: "abc",
    },
  };
  websocketStore.usersInRoom = [
    { name: "test po", role: "product-owner" },
    { name: "test dev", role: "developer", guess: 0 },
  ];
  websocketStore.possibleGuesses = [
    { guess: 1, description: "Bis zu 4 Std." },
    { guess: 2, description: "Bis zu 8 Std." },
    { guess: 3, description: "Bis zu 3 Tagen" },
    { guess: 4, description: "Bis zu 5 Tagen" },
    { guess: 5, description: "Mehr als 5 Tage" },
  ];
});
describe("RoomView", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      expect(wrapper.findComponent(VToolbar).exists()).to.be.true;
      expect(wrapper.findComponent(VToolbar).props("rounded")).to.be.true;

      expect(wrapper.findComponent(RoomDetail).exists()).to.be.true;
      expect(wrapper.findComponent(RoomDetail).props("usersInRoom")).equal(
        websocketStore.usersInRoom,
      );
      expect(wrapper.findComponent(RoomDetail).props("userRole")).equal(websocketStore.userRole);
      expect(wrapper.findComponent(RoomDetail).props("roundState")).equal(
        websocketStore.roundState,
      );
      expect(wrapper.findComponent(RoomDetail).props("ticketToGuess")).equal(
        websocketStore.ticketToGuess,
      );
      expect(wrapper.findComponent(RoomDetail).props("guess")).equal(websocketStore.guess);
      expect(wrapper.findComponent(RoomDetail).props("showAllGuesses")).equal(
        websocketStore.showAllGuesses,
      );
    });
  });

  describe("functionality", () => {
    it("should fetch possible guesses on mounted", async () => {
      vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      await nextTick();
      await nextTick();

      expect(websocketStore.fetchPossibleGuesses).toHaveBeenCalledOnce();
    });

    it("should render room form when user is not connected to the room", async () => {
      websocketStore.isConnected = false;
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      expect(wrapper.findComponent(RoomDetail).exists()).to.be.false;
      expect(wrapper.findComponent(RoomForm).exists()).to.be.true;
      expect(wrapper.findComponent(RoomForm).props("title")).equal("Join room");
      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal("");
      expect(wrapper.findComponent(RoomForm).props("showPasswordInput")).to.be.false;
      expect(wrapper.findComponent(RoomForm).props("subtitle")).equal(
        "You are currently not connected to this room",
      );
    });

    it("should render room form when user is not connected to the room and password is required", () => {
      websocketStore.isConnected = false;
      websocketStore.roomIsLocked = true;
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      expect(wrapper.findComponent(RoomForm).exists()).to.be.true;
      expect(wrapper.findComponent(RoomForm).props("showPasswordInput")).to.be.true;
    });

    it("should send correct message on estimate", () => {
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.findComponent(RoomDetail).vm.$emit("estimate", "WR-12");
      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "estimate",
        data: "WR-12",
      });
    });

    it("should send correct message on guess", () => {
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.findComponent(RoomDetail).vm.$emit("guess", 2);
      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "guess",
        data: 2,
      });
    });

    it("should send correct message on reveal", () => {
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.findComponent(RoomDetail).vm.$emit("reveal");
      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "reveal",
        data: null,
      });
    });

    it("should redirect to '/' when room does not exists", async () => {
      websocketStore.roomExists = vi.fn(() => Promise.resolve(false));
      vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      await nextTick();

      expect(useRouter().push).toHaveBeenNthCalledWith(1, "/");

      expect(websocketStore.fetchRoomIsLocked).not.toHaveBeenCalled();
    });

    it("should send correct message on new-round", () => {
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.findComponent(RoomDetail).vm.$emit("new-round");
      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "new-round",
        data: null,
      });
    });

    it("should disconnect and push to home when on leave", () => {
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper
        .findComponent(VToolbar)
        .findAllComponents(VBtn)
        .find((btn) => btn.findComponent(VIcon).find("i").classes().includes("mdi-location-exit"))
        .trigger("click");

      expect(websocketStore.disconnect).toHaveBeenCalledOnce();
      expect(useRouter().push).toHaveBeenNthCalledWith(1, "/");
    });

    it("should send correct message when locking room", async () => {
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      await wrapper
        .findComponent(VToolbar)
        .findAllComponents(VBtn)
        .find((btn) => btn.findComponent(VIcon).find("i").classes().includes("mdi-lock"))
        .trigger("click");

      expect(wrapper.findComponent(VDialog).exists()).to.be.true;

      await wrapper.findComponent(VDialog).findComponent(VTextField).setValue("top secret");
      await wrapper
        .findComponent(VDialog)
        .findAllComponents(VBtn)
        .find((btn) => btn.text() === "Lock")
        .trigger("click");

      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "lock-room",
        data: { password: "top secret", key: "abc" },
      });
    });

    it("should send correct message when opening room", async () => {
      websocketStore.roomIsLocked = true;
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      await wrapper
        .findComponent(VToolbar)
        .findAllComponents(VBtn)
        .find((btn) => btn.findComponent(VIcon).find("i").classes().includes("mdi-key"))
        .trigger("click");

      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "open-room",
        data: { key: "abc" },
      });
    });

    it("should copy password when room is locked and user has permissions", async () => {
      websocketStore.roomIsLocked = true;
      Object.defineProperty(global.navigator, "clipboard", {
        writable: true,
        value: {
          writeText: vi.fn(),
        },
      });
      Object.defineProperty(global.navigator, "permissions", {
        writable: true,
        value: {
          query: vi.fn().mockResolvedValue({ state: "granted" }),
        },
      });

      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      // @ts-ignore
      wrapper.vm.roomPassword = "top secret";
      await wrapper
        .findComponent(VToolbar)
        .findAllComponents(VBtn)
        .find((btn) => btn.findComponent(VIcon).find("i").classes().includes("mdi-content-copy"))
        .trigger("click");
      await nextTick();
      // @ts-ignore
      expect(wrapper.vm.snackbarText).equal("Copied!");
      expect(global.navigator.clipboard.writeText).toHaveBeenNthCalledWith(1, "top secret");
    });

    it("should not copy password when room is locked but no permission is granted", async () => {
      websocketStore.roomIsLocked = true;
      Object.defineProperty(global.navigator, "clipboard", {
        writable: true,
        value: {
          writeText: vi.fn(),
        },
      });
      Object.defineProperty(global.navigator, "permissions", {
        writable: true,
        value: {
          query: vi.fn().mockResolvedValue({ state: "denied" }),
        },
      });

      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      // @ts-ignore
      wrapper.vm.roomPassword = "top secret";
      await wrapper
        .findComponent(VToolbar)
        .findAllComponents(VBtn)
        .find((btn) => btn.findComponent(VIcon).find("i").classes().includes("mdi-content-copy"))
        .trigger("click");
      await nextTick();
      // @ts-ignore
      expect(wrapper.vm.snackbarText).equal("Could not copy");
      expect(global.navigator.clipboard.writeText).not.toHaveBeenCalled();
    });

    it("should join when user not connected and everything is fine", async () => {
      websocketStore.isConnected = false;
      websocketStore.isRoundInRoomInProgress = vi.fn(() => Promise.resolve(false));
      websocketStore.userExistsInRoom = vi.fn(() => Promise.resolve(false));
      websocketStore.connect = vi.fn();
      (useRoute as Mock).mockReturnValue({
        params: {
          id: "room-id",
        },
      });
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.findComponent(RoomForm).vm.$emit("update:name", "Name");
      wrapper.findComponent(RoomForm).vm.$emit("update:role", Role.Developer);
      wrapper.findComponent(RoomForm).vm.$emit("submit");

      await nextTick();
      await nextTick();
      await nextTick();

      expect(websocketStore.passwordMatchesRoom).not.toHaveBeenCalled();
      expect(websocketStore.isRoundInRoomInProgress).toHaveBeenNthCalledWith(1, "room-id");
      expect(websocketStore.userExistsInRoom).toHaveBeenNthCalledWith(1, "Name", "room-id");
      expect(websocketStore.connect).toHaveBeenNthCalledWith(1, "Name", Role.Developer, "room-id");
      expect(websocketStore.fetchPossibleGuesses).toHaveBeenCalledOnce();
      expect(websocketStore.fetchPermissions).toHaveBeenCalledOnce();
    });

    it("should not join when user is not connected, room is locked and password does not match", async () => {
      websocketStore.isConnected = false;
      websocketStore.roomIsLocked = true;
      websocketStore.passwordMatchesRoom = vi.fn(() => Promise.resolve(false));
      websocketStore.connect = vi.fn();
      (useRoute as Mock).mockReturnValue({
        params: {
          id: "room-id",
        },
      });
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.findComponent(RoomForm).vm.$emit("update:name", "Name");
      wrapper.findComponent(RoomForm).vm.$emit("update:role", Role.Developer);
      wrapper.findComponent(RoomForm).vm.$emit("update:password", "incorrect");
      wrapper.findComponent(RoomForm).vm.$emit("submit");

      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal(
        "The provided password is wrong",
      );

      expect(websocketStore.passwordMatchesRoom).toHaveBeenNthCalledWith(1, "room-id", "incorrect");
      expect(websocketStore.isRoundInRoomInProgress).not.toHaveBeenCalled();
      expect(websocketStore.userExistsInRoom).not.toHaveBeenCalled();
      expect(websocketStore.connect).not.toHaveBeenCalled();
      expect(websocketStore.fetchPossibleGuesses).not.toHaveBeenCalled();
      expect(websocketStore.fetchPermissions).not.toHaveBeenCalled();
    });

    it("should not join when user is not connected, password matches but round already started", async () => {
      websocketStore.isConnected = false;
      websocketStore.roomIsLocked = true;
      websocketStore.passwordMatchesRoom = vi.fn(() => Promise.resolve(true));
      websocketStore.isRoundInRoomInProgress = vi.fn(() => Promise.resolve(true));
      websocketStore.connect = vi.fn();
      (useRoute as Mock).mockReturnValue({
        params: {
          id: "room-id",
        },
      });
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.findComponent(RoomForm).vm.$emit("update:name", "Name");
      wrapper.findComponent(RoomForm).vm.$emit("update:role", Role.Developer);
      wrapper.findComponent(RoomForm).vm.$emit("update:password", "correct");
      wrapper.findComponent(RoomForm).vm.$emit("submit");

      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal(
        "The round has already started",
      );

      expect(websocketStore.passwordMatchesRoom).toHaveBeenNthCalledWith(1, "room-id", "correct");
      expect(websocketStore.isRoundInRoomInProgress).toHaveBeenNthCalledWith(1, "room-id");
      expect(websocketStore.userExistsInRoom).not.toHaveBeenCalled();
      expect(websocketStore.connect).not.toHaveBeenCalled();
      expect(websocketStore.fetchPossibleGuesses).not.toHaveBeenCalled();
      expect(websocketStore.fetchPermissions).not.toHaveBeenCalled();
    });

    it("should not join when user is not connected, password matches but user already exists in the room", async () => {
      websocketStore.isConnected = false;
      websocketStore.roomIsLocked = true;
      websocketStore.passwordMatchesRoom = vi.fn(() => Promise.resolve(true));
      websocketStore.isRoundInRoomInProgress = vi.fn(() => Promise.resolve(false));
      websocketStore.userExistsInRoom = vi.fn(() => Promise.resolve(true));
      websocketStore.connect = vi.fn();
      (useRoute as Mock).mockReturnValue({
        params: {
          id: "room-id",
        },
      });
      const wrapper = vuetifyMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.findComponent(RoomForm).vm.$emit("update:name", "Name");
      wrapper.findComponent(RoomForm).vm.$emit("update:role", Role.Developer);
      wrapper.findComponent(RoomForm).vm.$emit("update:password", "correct");
      wrapper.findComponent(RoomForm).vm.$emit("submit");

      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal(
        "A user with this name already exists in the room",
      );

      expect(websocketStore.passwordMatchesRoom).toHaveBeenNthCalledWith(1, "room-id", "correct");
      expect(websocketStore.isRoundInRoomInProgress).toHaveBeenNthCalledWith(1, "room-id");
      expect(websocketStore.userExistsInRoom).toHaveBeenNthCalledWith(1, "Name", "room-id");
      expect(websocketStore.connect).not.toHaveBeenCalled();
      expect(websocketStore.fetchPossibleGuesses).not.toHaveBeenCalled();
      expect(websocketStore.fetchPermissions).not.toHaveBeenCalled();
    });
  });
});
