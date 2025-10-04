import { beforeEach, describe, it, expect, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import HomeView from "../../src/views/HomeView.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { useWebsocketStore } from "../../src/stores/websocket";
import { vuetifyMount } from "../vuetifyMount";
import { VCard, VIcon } from "vuetify/components";
import RoomDialog from "../../src/components/RoomDialog.vue";
import { Role } from "../../src/components/types";
import { nextTick } from "vue";

vi.mock("vue-router");

let pinia: TestingPinia;
let websocketStore: ReturnType<typeof useWebsocketStore>;

beforeEach(() => {
  pinia = createTestingPinia();
  websocketStore = useWebsocketStore(pinia);
  websocketStore.fetchActiveRooms = vi.fn(() => Promise.resolve([]));
});
describe("HomeView", () => {
  describe("rendering", () => {
    it("should render without rooms", () => {
      const wrapper = vuetifyMount(HomeView, {
        global: {
          plugins: [pinia],
        },
      });

      expect(wrapper.findComponent(VCard).exists()).to.be.false;
      expect(wrapper.findComponent(VIcon).exists()).to.be.true;
      expect(wrapper.findComponent(VIcon).props("icon")).equal("mdi-magnify");
      expect(wrapper.findComponent(VIcon).props("size")).equal("80");
      expect(wrapper.findComponent(VIcon).classes()).contains("opacity-50");

      expect(wrapper.find("span").text()).equal("There are currently no rooms");
      expect(wrapper.find("span").classes()).deep.equal(["text-h4", "opacity-90"]);

      expect(wrapper.findComponent(RoomDialog).exists()).to.be.true;
      expect(wrapper.findComponent(RoomDialog).props("role")).equal(Role.Empty);
      expect(wrapper.findComponent(RoomDialog).props("name")).equal("");
      expect(wrapper.findComponent(RoomDialog).props("activatorText")).equal("Create a new one");
      expect(wrapper.findComponent(RoomDialog).props("cardTitle")).equal("Create room");
      expect(wrapper.findComponent(RoomDialog).props("errorMessage")).to.be.undefined;
    });

    it("should render with rooms", async () => {
      websocketStore.fetchActiveRooms = vi.fn(() =>
        Promise.resolve([
          {
            id: "first-id",
            playerCount: 1,
          },
          {
            id: "second-id",
            playerCount: 3,
          },
        ]),
      );

      const wrapper = vuetifyMount(HomeView, {
        global: {
          plugins: [pinia],
        },
      });

      await nextTick();

      expect(wrapper.findComponent(VCard).exists()).to.be.true;
      expect(wrapper.findAllComponents(VCard)).length(2);
      wrapper.findAllComponents(VCard).forEach((el, index) => {
        expect(el.props("title")).equal(`Room #${index + 1}`);
        expect(el.props("variant")).equal("outlined");
        expect(el.props("prependIcon")).equal("mdi-poker-chip");
        expect(el.props("maxWidth")).equal("450");
      });
      expect(wrapper.findAllComponents(VCard).at(0).props("subtitle")).equal("first-id");
      expect(wrapper.findAllComponents(VCard).at(1).props("subtitle")).equal("second-id");
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
