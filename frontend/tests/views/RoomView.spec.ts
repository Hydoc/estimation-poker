import { beforeEach, describe, expect, it, Mock, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import RoomView from "../../src/views/RoomView.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { useWebsocketStore } from "../../src/stores/websocket";
import { useRouter } from "vue-router";
import RoomDetail from "../../src/components/RoomDetail.vue";
import { RoundState } from "../../src/components/types";

vi.mock("vue-router");

let pinia: TestingPinia;
let websocketStore: ReturnType<typeof useWebsocketStore>;

beforeEach(() => {
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
  websocketStore.usersInRoom = {
    productOwnerList: [{ name: "test po", role: "product-owner" }],
    developerList: [{ name: "test dev", role: "developer", guess: 0 }],
  };
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
          plugins: [pinia],
        },
      });

      expect(wrapper.findComponent(RoomDetail).exists()).to.be.true;
      expect(wrapper.findComponent(RoomDetail).props("currentUsername")).equal(
        websocketStore.username,
      );
      expect(wrapper.findComponent(RoomDetail).props("roomId")).equal(websocketStore.roomId);
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
          plugins: [pinia],
        },
      });

      expect(websocketStore.fetchPossibleGuesses).toHaveBeenCalledOnce();
    });

    it("should push to home when user is not connected", () => {
      websocketStore.isConnected = false;
      shallowMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      expect(useRouter().push).toHaveBeenNthCalledWith(1, "/");
    });

    it("should send correct message on estimate", () => {
      const wrapper = shallowMount(RoomView, {
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
      const wrapper = shallowMount(RoomView, {
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
      const wrapper = shallowMount(RoomView, {
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

    it("should send correct message on new-round", () => {
      const wrapper = shallowMount(RoomView, {
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
      const wrapper = shallowMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.findComponent(RoomDetail).vm.$emit("leave");

      expect(websocketStore.disconnect).toHaveBeenCalledOnce();
      expect(useRouter().push).toHaveBeenNthCalledWith(1, "/");
    });

    it("should send correct message on lock-room", () => {
      const wrapper = shallowMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper
        .findComponent(RoomDetail)
        .vm.$emit("lock-room", { password: "top secret", key: "abc" });
      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "lock-room",
        data: { password: "top secret", key: "abc" },
      });
    });

    it("should send correct message on open-room", () => {
      const wrapper = shallowMount(RoomView, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.findComponent(RoomDetail).vm.$emit("open-room", { key: "abc" });
      expect(websocketStore.send).toHaveBeenNthCalledWith(1, {
        type: "open-room",
        data: { key: "abc" },
      });
    });
  });
});
