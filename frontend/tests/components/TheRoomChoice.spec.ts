import { describe, it, expect, beforeEach, vi, Mock } from "vitest";
import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { useRouter } from "vue-router";
import TheRoomChoice from "../../src/components/TheRoomChoice.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { VAlert, VBtn, VCard, VForm, VRadio, VRadioGroup, VTextField } from "vuetify/components";
import { Role } from "../../src/components/types";
import { nextTick } from "vue";
import { useWebsocketStore } from "../../src/stores/websocket";

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

  (useRouter as Mock).mockReturnValue({
    push: vi.fn(),
  });
});
describe("TheRoomChoice", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, createTestingPinia()],
        },
      });

      expect(wrapper.findComponent(VCard).exists()).to.be.true;
      expect(wrapper.findComponent(VCard).props("prependIcon")).equal("mdi-poker-chip");
      expect(wrapper.findComponent(VCard).text()).contains(
        "Ich brauche noch ein paar Informationen bevor es los geht",
      );

      expect(wrapper.findComponent(VForm).exists()).to.be.true;
      expect(wrapper.findComponent(VForm).props("fastFail")).to.be.true;
      expect(wrapper.findComponent(VForm).props("validateOn")).equal("input");

      expect(wrapper.findAllComponents(VTextField).at(0).props("label")).equal("Raum");
      expect(
        Object.keys(wrapper.findAllComponents(VTextField).at(0).find("input").attributes()),
      ).contains("required");
      expect(wrapper.findAllComponents(VTextField).at(1).props("label")).equal("Name");
      expect(
        Object.keys(wrapper.findAllComponents(VTextField).at(0).find("input").attributes()),
      ).contains("required");

      expect(wrapper.findComponent(VRadioGroup).exists()).to.be.true;
      expect(wrapper.findComponent(VRadioGroup).props("label")).equal("Deine Rolle");
      expect(
        wrapper.findComponent(VRadioGroup).findAllComponents(VRadio).at(0).props("label"),
      ).equal("Product Owner");
      expect(
        wrapper.findComponent(VRadioGroup).findAllComponents(VRadio).at(0).props("value"),
      ).equal(Role.ProductOwner);
      expect(
        wrapper.findComponent(VRadioGroup).findAllComponents(VRadio).at(1).props("label"),
      ).equal("Entwickler");
      expect(
        wrapper.findComponent(VRadioGroup).findAllComponents(VRadio).at(1).props("value"),
      ).equal(Role.Developer);

      expect(wrapper.findComponent(VAlert).exists()).to.be.false;

      expect(wrapper.findComponent(VBtn).exists()).to.be.true;
      expect(wrapper.findComponent(VBtn).text()).equal("Verbinden");
      expect(wrapper.findComponent(VBtn).find("button").attributes("type")).equal("submit");
      expect(wrapper.findComponent(VBtn).props("color")).equal("primary");
      expect(wrapper.findComponent(VBtn).props("prependIcon")).equal("mdi-connection");
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;
    });
  });

  describe("functionality", () => {
    it("should enable button when everything is valid", async () => {
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, createTestingPinia()],
        },
      });
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;

      await wrapper.findAllComponents(VTextField).at(0).setValue("Blub");
      await wrapper.findAllComponents(VTextField).at(1).setValue("My name");
      await wrapper.findComponent(VRadioGroup).setValue(Role.Developer);

      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.false;
    });

    it("should show validation messages when fields are cleared", async () => {
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, createTestingPinia()],
        },
      });

      await wrapper.findAllComponents(VTextField).at(0).setValue("test");
      await wrapper.findAllComponents(VTextField).at(0).setValue("");
      await nextTick();
      expect(wrapper.findAllComponents(VTextField).at(0).text()).contains(
        "Fehler: Hier müsste eigentlich was stehen",
      );

      await wrapper.findAllComponents(VTextField).at(1).setValue("test 2");
      await wrapper.findAllComponents(VTextField).at(1).setValue("");
      await nextTick();
      expect(wrapper.findAllComponents(VTextField).at(1).text()).contains(
        "Fehler: Hier müsste eigentlich was stehen",
      );
    });

    it("should connect when connect is clicked with everything valid", async () => {
      const wrapper = mount(TheRoomChoice, {
        global: {
          plugins: [vuetify, pinia],
        },
      });

      await wrapper.findAllComponents(VTextField).at(0).setValue("test");
      await wrapper.findAllComponents(VTextField).at(1).setValue("my name");
      await wrapper.findComponent(VRadioGroup).setValue(Role.Developer);
      await wrapper.findComponent(VBtn).trigger("submit");
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

      await wrapper.findAllComponents(VTextField).at(0).setValue("test");
      await wrapper.findAllComponents(VTextField).at(1).setValue("my name");
      await wrapper.findComponent(VRadioGroup).setValue(Role.Developer);
      await wrapper.findComponent(VBtn).trigger("submit");

      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(VAlert).exists()).to.be.true;
      expect(wrapper.findComponent(VAlert).props("color")).equal("error");
      expect(wrapper.findComponent(VAlert).props("text")).equal(
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

      await wrapper.findAllComponents(VTextField).at(0).setValue("test");
      await wrapper.findAllComponents(VTextField).at(1).setValue("my name");
      await wrapper.findComponent(VRadioGroup).setValue(Role.Developer);
      await wrapper.findComponent(VBtn).trigger("submit");

      await nextTick();
      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(VAlert).exists()).to.be.true;
      expect(wrapper.findComponent(VAlert).props("color")).equal("error");
      expect(wrapper.findComponent(VAlert).props("text")).equal(
        "Die Runde in diesem Raum hat bereits begonnen.",
      );

      expect(useRouter().push).not.toHaveBeenCalled();
      expect(websocketStore.isRoundInRoomInProgress).toHaveBeenNthCalledWith(1, "test");
      expect(websocketStore.connect).not.toHaveBeenCalled();
    });
  });
});
