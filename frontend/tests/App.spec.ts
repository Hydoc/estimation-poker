import { beforeEach, describe, expect, it, vi } from "vitest";
import { mount, shallowMount } from "@vue/test-utils";
import App from "../src/App.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { useWebsocketStore } from "../src/stores/websocket";
import { RouterView } from "vue-router";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { createVuetify } from "vuetify";
import { VApp, VAppBar, VAppBarTitle } from "vuetify/components";

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
beforeEach(() => {
  pinia = createTestingPinia();
  websocketStore = useWebsocketStore(pinia);
  vuetify = createVuetify({
    components,
    directives,
  });
});
describe("App", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = mount(App, {
        global: {
          plugins: [pinia, vuetify],
          stubs: {
            RouterView: {
              render() {},
            },
          },
        },
      });

      expect(wrapper.findComponent(VApp).exists()).to.be.true;
      expect(wrapper.findComponent(VAppBar).exists()).to.be.true;
      expect(wrapper.findComponent(VAppBar).findComponent(VAppBarTitle).text()).equal(
        "Estimation Poker",
      );
      expect(wrapper.findComponent(RouterView).exists()).to.be.true;
    });
  });

  describe("functionality", () => {
    it("should disconnect from websocket when unmounting", () => {
      const wrapper = shallowMount(App, {
        global: {
          plugins: [pinia, vuetify],
        },
      });

      wrapper.unmount();
      expect(websocketStore.disconnect).toHaveBeenCalledOnce();
    });
  });
});
