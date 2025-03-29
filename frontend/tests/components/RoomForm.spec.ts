import { describe, it, expect, beforeEach } from "vitest";
import { Role } from "../../src/components/types";
import { mount } from "@vue/test-utils";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { createVuetify } from "vuetify";
import RoomForm from "../../src/components/RoomForm.vue";
import { VAlert, VBtn, VForm, VRadio, VRadioGroup, VTextField } from "vuetify/components";
import { nextTick } from "vue";

let vuetify: ReturnType<typeof createVuetify>;

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});
describe("RoomForm", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = mount(RoomForm, {
        props: {
          name: "",
          roomId: "",
          role: Role.Empty,
        },
        global: {
          plugins: [vuetify],
        },
      });

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
      const wrapper = mount(RoomForm, {
        props: {
          name: "",
          roomId: "",
          role: Role.Empty,
        },
        global: {
          plugins: [vuetify],
        },
      });
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;

      await wrapper.findAllComponents(VTextField).at(0).setValue("Blub");
      await wrapper.findAllComponents(VTextField).at(1).setValue("My name");
      await wrapper.findComponent(VRadioGroup).setValue(Role.Developer);
      
      await nextTick();

      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.false;
    });

    it("should show validation messages when fields are cleared", async () => {
      const wrapper = mount(RoomForm, {
        props: {
          name: "",
          roomId: "",
          role: Role.Empty,
        },
        global: {
          plugins: [vuetify],
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
  });
});
