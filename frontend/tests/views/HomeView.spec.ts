import { beforeEach, describe, it, expect } from "vitest";
import { shallowMount } from "@vue/test-utils";
import HomeView from "../../src/views/HomeView.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { useWebsocketStore } from "../../src/stores/websocket";
import TheRoomChoice from "../../src/components/TheRoomChoice.vue";

let pinia: TestingPinia;
let websocketStore: ReturnType<typeof useWebsocketStore>;

beforeEach(() => {
  pinia = createTestingPinia();
  websocketStore = useWebsocketStore(pinia);
});
describe("HomeView", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = shallowMount(HomeView, {
        global: {
          plugins: [pinia],
        },
      });

      expect(wrapper.findComponent(TheRoomChoice).exists()).to.be.true;
    });
  });

  describe("functionality", () => {
    it("should reset round and close websocket on render", () => {
      shallowMount(HomeView, {
        global: {
          plugins: [pinia],
        },
      });

      expect(websocketStore.disconnect).toHaveBeenCalledOnce();
      expect(websocketStore.resetRound).toHaveBeenCalledOnce();
    });
  });
});
