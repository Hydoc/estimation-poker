import { beforeEach, describe, expect, it, Mock, vi } from "vitest";
import HomeView from "../../src/views/HomeView.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { vuetifyMount } from "../vuetifyMount";
import { VCard, VCardText, VIcon } from "vuetify/components";
import RoomDialog from "../../src/components/RoomDialog.vue";
import { nextTick } from "vue";
import { useRouter } from "vue-router";
import { useEstimationStore } from "../../src/stores/estimation";
import { Role } from "../../src/types/room";
import { succeed } from "@kaumlaut/pure/fetch-state";

vi.mock("vue-router");

let pinia: TestingPinia;
let estimationStore: ReturnType<typeof useEstimationStore>;

beforeEach(() => {
  (useRouter as Mock).mockReturnValue({
    push: vi.fn(),
  });

  pinia = createTestingPinia();
  estimationStore = useEstimationStore(pinia);
  // @ts-ignore
  estimationStore.fetchActiveRooms = vi.fn(() => Promise.resolve([]));
  // @ts-ignore
  estimationStore.joinRoom = vi.fn();
});
describe("HomeView", () => {
  describe("rendering", () => {
    it("should render without rooms", () => {
      const wrapper = createWrapper();

      expect(wrapper.findComponent(VCard).exists()).to.be.false;
      expect(wrapper.findComponent(VIcon).exists()).to.be.true;
      expect(wrapper.findComponent(VIcon).props("icon")).equal("mdi-magnify");
      expect(wrapper.findComponent(VIcon).props("size")).equal("80");
      expect(wrapper.findComponent(VIcon).classes()).contains("opacity-50");

      expect(wrapper.find("span").text()).equal("There are currently no rooms");
      expect(wrapper.find("span").classes()).deep.equal(["text-h4", "opacity-90"]);

      expect(wrapper.findComponent(RoomDialog).exists()).to.be.true;
      expect(wrapper.findAllComponents(RoomDialog)).length(1);
      expect(wrapper.findComponent(RoomDialog).props("role")).equal(Role.Empty);
      expect(wrapper.findComponent(RoomDialog).props("name")).equal("");
      expect(wrapper.findComponent(RoomDialog).props("activatorText")).equal("Create a new one");
      expect(wrapper.findComponent(RoomDialog).props("cardTitle")).equal("Create room");
      expect(wrapper.findComponent(RoomDialog).props("errorMessage")).to.be.undefined;
    });

    it("should render with rooms", async () => {
      // @ts-ignore
      estimationStore.roomsState.availableActiveRooms = succeed({
        rooms: [
          {
            id: "first-id",
            playerCount: 1,
          },
          {
            id: "second-id",
            playerCount: 3,
          },
        ],
      });

      const wrapper = createWrapper();

      await nextTick();

      expect(wrapper.findAllComponents(RoomDialog)).length(3);
      expect(wrapper.findAllComponents(RoomDialog).at(0).props("activatorText")).equal(
        "Create a new room",
      );
      expect(wrapper.findAllComponents(RoomDialog).at(1).props("activatorText")).equal("Join");
      expect(wrapper.findAllComponents(RoomDialog).at(2).props("activatorText")).equal("Join");
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

      expect(wrapper.findAllComponents(VCardText).at(0).findComponent(VIcon).props("icon")).equal(
        "mdi-account",
      );
      expect(wrapper.findAllComponents(VCardText).at(0).text()).equal("1 player");
      expect(wrapper.findAllComponents(VCardText).at(1).findComponent(VIcon).props("icon")).equal(
        "mdi-account",
      );
      expect(wrapper.findAllComponents(VCardText).at(1).text()).equal("3 players");
    });
  });

  describe("functionality", () => {
    it("should reset round and close websocket on render", () => {
      createWrapper();

      // @ts-ignore
      expect(estimationStore.leaveRoom).toHaveBeenCalledOnce();
    });

    it("should create a new room", async () => {
      // @ts-ignore
      estimationStore.createRoom = vi.fn(() => Promise.resolve("room-id"));
      // @ts-ignore
      estimationStore.fetchRoomState = vi.fn(() =>
        Promise.resolve({
          isLocked: false,
          inProgress: false,
        }),
      );
      // @ts-ignore
      estimationStore.userExists = vi.fn(() => Promise.resolve(false));

      const wrapper = createWrapper();

      wrapper.findComponent(RoomDialog).vm.$emit("update:name", "Name");
      wrapper.findComponent(RoomDialog).vm.$emit("update:role", Role.Developer);
      wrapper.findComponent(RoomDialog).vm.$emit("submit");

      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      // @ts-ignore
      expect(estimationStore.createRoom).toHaveBeenNthCalledWith(1, "Name");
      // @ts-ignore
      expect(estimationStore.userExists).toHaveBeenNthCalledWith(1, "room-id", "Name");
      // @ts-ignore
      expect(estimationStore.joinRoom).toHaveBeenNthCalledWith(
        1,
        "Name",
        Role.Developer,
        "room-id",
      );
      expect(useRouter().push).toHaveBeenNthCalledWith(1, "/room/room-id");
    });
  });
});

function createWrapper() {
  return vuetifyMount(HomeView, {
    global: {
      plugins: [pinia],
    },
  });
}
