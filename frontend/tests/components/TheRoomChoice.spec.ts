import { describe, it, expect, beforeEach, vi, Mock } from "vitest";
import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { useRouter } from "vue-router";
import TheRoomChoice from "../../src/components/TheRoomChoice.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { VCard } from "vuetify/components";
import { Role } from "../../src/components/types";
import { nextTick } from "vue";
import { useWebsocketStore } from "../../src/stores/websocket";
import RoomForm from "../../src/components/RoomForm.vue";
import TheActiveRoomOverview from "../../src/components/TheActiveRoomOverview.vue";

vi.mock("vue-router");

let vuetify: ReturnType<typeof createVuetify>;
let pinia: TestingPinia;
let websocketStore: ReturnType<typeof useWebsocketStore>;

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });

  pinia = createTestingPinia();
  websocketStore = useWebsocketStore(pinia);
  websocketStore.userExistsInRoom = vi.fn().mockResolvedValue(false);
  websocketStore.isRoundInRoomInProgress = vi.fn().mockResolvedValue(false);
  websocketStore.fetchActiveRooms = vi.fn().mockResolvedValue([]);

  (useRouter as Mock).mockReturnValue({
    push: vi.fn(),
  });
});
describe("TheRoomChoice", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, pinia],
        },
      });

      expect(wrapper.findComponent(VCard).exists()).to.be.true;
      expect(wrapper.findComponent(VCard).props("prependIcon")).equal("mdi-poker-chip");
      expect(wrapper.findComponent(VCard).text()).contains(
        "Ich brauche noch ein paar Informationen bevor es los geht",
      );

      expect(wrapper.findComponent(RoomForm).exists()).to.be.true;
      expect(wrapper.findComponent(TheActiveRoomOverview).exists()).to.be.false;
    });

    it("should render room overview when active rooms are found", async () => {
      websocketStore.fetchActiveRooms = vi.fn().mockResolvedValue(["Hello"]);
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, pinia],
        },
      });
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(TheActiveRoomOverview).exists()).to.be.true;
      expect(wrapper.findComponent(TheActiveRoomOverview).props("activeRooms")).deep.equal([
        "Hello",
      ]);
    });
  });

  describe("functionality", () => {
    it("should connect when submit was emitted", async () => {
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, pinia],
        },
      });

      wrapper.vm.role = Role.Developer;
      wrapper.vm.name = "my name";
      wrapper.vm.roomId = "test";

      await wrapper.findComponent(RoomForm).vm.$emit("submit");
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      expect(useRouter().push).toHaveBeenNthCalledWith(1, "/room");
      expect(websocketStore.userExistsInRoom).toHaveBeenNthCalledWith(1, "my name", "test");
      expect(websocketStore.connect).toHaveBeenNthCalledWith(1, "my name", "developer", "test");
    });

    it("should show error when user in room already exists", async () => {
      websocketStore.userExistsInRoom = vi.fn().mockResolvedValue(true);
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, pinia],
        },
      });

      wrapper.vm.role = Role.Developer;
      wrapper.vm.name = "my name";
      wrapper.vm.roomId = "test";
      await wrapper.findComponent(RoomForm).vm.$emit("submit");
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal(
        "Ein Benutzer mit diesem Namen existiert in dem Raum bereits.",
      );

      expect(useRouter().push).not.toHaveBeenCalled();
      expect(websocketStore.userExistsInRoom).toHaveBeenNthCalledWith(1, "my name", "test");
      expect(websocketStore.connect).not.toHaveBeenCalled();
    });

    it("should show error when round in room is in progress", async () => {
      websocketStore.isRoundInRoomInProgress = vi.fn().mockResolvedValue(true);
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, pinia],
        },
      });

      wrapper.vm.role = Role.Developer;
      wrapper.vm.name = "my name";
      wrapper.vm.roomId = "test";
      await wrapper.findComponent(RoomForm).vm.$emit("submit");
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal(
        "Die Runde in diesem Raum hat bereits begonnen.",
      );

      expect(useRouter().push).not.toHaveBeenCalled();
      expect(websocketStore.isRoundInRoomInProgress).toHaveBeenNthCalledWith(1, "test");
      expect(websocketStore.connect).not.toHaveBeenCalled();
    });

    it("should join when room overview emits join", async () => {
      websocketStore.fetchActiveRooms = vi.fn().mockResolvedValue(["Hello"]);
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, pinia],
        },
      });
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      await wrapper
        .findComponent(TheActiveRoomOverview)
        .vm.$emit("join", "test", "my name", Role.Developer);
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      expect(useRouter().push).toHaveBeenNthCalledWith(1, "/room");
      expect(websocketStore.userExistsInRoom).toHaveBeenNthCalledWith(1, "my name", "test");
      expect(websocketStore.connect).toHaveBeenNthCalledWith(1, "my name", "developer", "test");
    });

    it("should show error when trying to join but user in room already exists", async () => {
      websocketStore.userExistsInRoom = vi.fn().mockResolvedValue(true);
      websocketStore.fetchActiveRooms = vi.fn().mockResolvedValue(["Hello"]);
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, pinia],
        },
      });
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      await wrapper
        .findComponent(TheActiveRoomOverview)
        .vm.$emit("join", "test", "my name", Role.Developer);
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(TheActiveRoomOverview).props("errorMessage")).equal(
        "Ein Benutzer mit diesem Namen existiert in dem Raum bereits.",
      );

      expect(useRouter().push).not.toHaveBeenCalled();
      expect(websocketStore.userExistsInRoom).toHaveBeenNthCalledWith(1, "my name", "test");
      expect(websocketStore.connect).not.toHaveBeenCalled();
    });

    it("should show error when round in room is in progress while trying to join", async () => {
      websocketStore.isRoundInRoomInProgress = vi.fn().mockResolvedValue(true);
      websocketStore.fetchActiveRooms = vi.fn().mockResolvedValue(["Hello"]);
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, pinia],
        },
      });
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      await wrapper
        .findComponent(TheActiveRoomOverview)
        .vm.$emit("join", "test", "my name", Role.Developer);
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(TheActiveRoomOverview).props("errorMessage")).equal(
        "Die Runde in diesem Raum hat bereits begonnen.",
      );

      expect(useRouter().push).not.toHaveBeenCalled();
      expect(websocketStore.isRoundInRoomInProgress).toHaveBeenNthCalledWith(1, "test");
      expect(websocketStore.connect).not.toHaveBeenCalled();
    });
  });
});
