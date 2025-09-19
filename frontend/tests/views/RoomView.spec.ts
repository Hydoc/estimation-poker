import { beforeEach, describe, expect, it, Mock, vi } from "vitest";
import { mount, shallowMount } from "@vue/test-utils";
import RoomView from "../../src/views/RoomView.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { useWebsocketStore } from "../../src/stores/websocket";
import { useRouter } from "vue-router";
import RoomDetail from "../../src/components/RoomDetail.vue";
import { Role, RoundState } from "../../src/components/types";
import { createVuetify } from "vuetify";
import { VBtn, VDialog, VIcon, VTextField, VToolbar } from "vuetify/components";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { nextTick } from "vue";

vi.mock("vue-router");

let pinia: TestingPinia;
let websocketStore: ReturnType<typeof useWebsocketStore>;
let vuetify: ReturnType<typeof createVuetify>;

const ResizeObserverMock = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));
vi.stubGlobal("ResizeObserver", ResizeObserverMock);
vi.stubGlobal("visualViewport", new EventTarget());
beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
  (useRouter as Mock).mockReturnValue({
    push: vi.fn(),
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
      const wrapper = shallowMount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
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
    it("should fetch possible guesses on mounted", () => {
      shallowMount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
        },
      });

      expect(websocketStore.fetchPossibleGuesses).toHaveBeenCalledOnce();
    });

    it("should push to home when user is not connected", () => {
      websocketStore.isConnected = false;
      shallowMount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
        },
      });

      expect(useRouter().push).toHaveBeenNthCalledWith(1, "/");
    });

    it("should send correct message on estimate", () => {
      const wrapper = shallowMount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
        },
      });

      wrapper.findComponent(RoomDetail).vm.$emit("estimate", "WR-12");
      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "estimate",
        data: "WR-12",
      });
    });

    it("should send correct message on guess", () => {
      const wrapper = shallowMount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
        },
      });

      wrapper.findComponent(RoomDetail).vm.$emit("guess", 2);
      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "guess",
        data: 2,
      });
    });

    it("should send correct message on reveal", () => {
      const wrapper = shallowMount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
        },
      });

      wrapper.findComponent(RoomDetail).vm.$emit("reveal");
      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "reveal",
        data: null,
      });
    });

    it("should send correct message on new-round", () => {
      const wrapper = shallowMount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
        },
      });

      wrapper.findComponent(RoomDetail).vm.$emit("new-round");
      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "new-round",
        data: null,
      });
    });

    it("should disconnect and push to home when on leave", () => {
      const wrapper = mount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
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
      const wrapper = mount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
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
        .find((btn) => btn.text() === "Abschließen")
        .trigger("click");

      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "lock-room",
        data: { password: "top secret", key: "abc" },
      });
    });

    it("should send correct message when opening room", async () => {
      websocketStore.roomIsLocked = true;
      const wrapper = mount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
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

      const wrapper = mount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
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
      expect(wrapper.vm.snackbarText).equal("Kopiert!");
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

      const wrapper = mount(RoomView, {
        global: {
          plugins: [pinia, vuetify],
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
      expect(wrapper.vm.snackbarText).equal("Konnte nicht kopiert werden");
      expect(global.navigator.clipboard.writeText).not.toHaveBeenCalled();
    });
  });
});
