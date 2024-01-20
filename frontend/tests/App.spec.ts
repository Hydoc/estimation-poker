import { beforeEach, describe, expect, it } from "vitest";
import { shallowMount } from "@vue/test-utils";
import App from "../src/App.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { useWebsocketStore } from "../src/stores/websocket";
import { RouterView } from "vue-router";

let pinia: TestingPinia;
let websocketStore: ReturnType<typeof useWebsocketStore>;

beforeEach(() => {
  pinia = createTestingPinia();
  websocketStore = useWebsocketStore(pinia);
});
describe("App", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = shallowMount(App, {
        global: {
          plugins: [pinia],
        },
      });

      expect(wrapper.find("h1").text()).equal("Estimation Poker");
      expect(wrapper.find("header").exists()).to.be.true;
      expect(wrapper.findComponent(RouterView).exists()).to.be.true;
    });
  });

  describe("functionality", () => {
    it("should disconnect from websocket when unmounting", () => {
      const wrapper = shallowMount(App, {
        global: {
          plugins: [pinia],
        },
      });

      wrapper.unmount();
      expect(websocketStore.disconnect).toHaveBeenCalledOnce();
    });
  });
});
