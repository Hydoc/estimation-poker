import { describe, expect, it } from "vitest";
import { Role } from "../../src/components/types";
import RoomForm from "../../src/components/RoomForm.vue";
import {
  VAlert,
  VBtn,
  VCardSubtitle,
  VForm,
  VRadio,
  VRadioGroup,
  VTextField,
} from "vuetify/components";
import { nextTick } from "vue";
import { vuetifyMount } from "../vuetifyMount";

const defaultProps = {
  subtitle: "Room form subtitle",
  title: "Room Form",
  name: "",
  role: Role.Empty,
};

describe("RoomForm", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = vuetifyMount(RoomForm, {
        props: defaultProps,
      });

      expect(wrapper.findComponent(VCardSubtitle).exists()).to.be.true;
      expect(wrapper.findComponent(VCardSubtitle).text()).equal("Room form subtitle");

      expect(wrapper.findComponent(VForm).exists()).to.be.true;
      expect(wrapper.findComponent(VForm).props("fastFail")).to.be.true;
      expect(wrapper.findComponent(VForm).props("validateOn")).equal("input");

      expect(wrapper.findAllComponents(VTextField).at(0).props("label")).equal("Name");
      expect(
        Object.keys(wrapper.findAllComponents(VTextField).at(0).find("input").attributes()),
      ).contains("required");

      expect(wrapper.findComponent(VRadioGroup).exists()).to.be.true;
      expect(wrapper.findComponent(VRadioGroup).props("label")).equal("Your role");
      expect(
        wrapper.findComponent(VRadioGroup).findAllComponents(VRadio).at(0).props("label"),
      ).equal("Product Owner");
      expect(
        wrapper.findComponent(VRadioGroup).findAllComponents(VRadio).at(0).props("value"),
      ).equal(Role.ProductOwner);
      expect(
        wrapper.findComponent(VRadioGroup).findAllComponents(VRadio).at(1).props("label"),
      ).equal("Developer");
      expect(
        wrapper.findComponent(VRadioGroup).findAllComponents(VRadio).at(1).props("value"),
      ).equal(Role.Developer);

      expect(wrapper.findComponent(VAlert).exists()).to.be.false;

      expect(wrapper.findComponent(VBtn).exists()).to.be.true;
      expect(wrapper.findComponent(VBtn).text()).equal("Connect");
      expect(wrapper.findComponent(VBtn).props("color")).equal("primary");
      expect(wrapper.findComponent(VBtn).props("prependIcon")).equal("mdi-connection");
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;
    });

    it("should render without subtitle", () => {
      const wrapper = vuetifyMount(RoomForm, {
        props: {
          ...defaultProps,
          subtitle: undefined,
        },
      });

      expect(wrapper.findComponent(VCardSubtitle).exists()).to.be.false;
    });

    it("should render alert when errorMessage given", () => {
      const wrapper = vuetifyMount(RoomForm, {
        props: {
          ...defaultProps,
          errorMessage: "Something went wrong",
        },
      });

      expect(wrapper.findComponent(VAlert).exists()).to.be.true;
      expect(wrapper.findComponent(VAlert).props("text")).equal("Something went wrong");
      expect(wrapper.findComponent(VAlert).props("color")).equal("error");
    });

    it("should render password input when showPasswordInput = true", () => {
      const wrapper = vuetifyMount(RoomForm, {
        props: {
          ...defaultProps,
          showPasswordInput: true,
        },
      });

      expect(wrapper.findAllComponents(VTextField)).length(2);
      expect(wrapper.findAllComponents(VTextField).at(1).props("label")).equal("Password");
      expect(wrapper.findAllComponents(VTextField).at(1).props("type")).equal("password");
      expect(
        Object.keys(wrapper.findAllComponents(VTextField).at(1).find("input").attributes()),
      ).contains("required");
    });
  });

  describe("functionality", () => {
    it("should enable button when everything is valid", async () => {
      const wrapper = vuetifyMount(RoomForm, {
        props: defaultProps,
      });
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;

      await wrapper.findAllComponents(VTextField).at(0).setValue("My name");
      await wrapper.findComponent(VRadioGroup).setValue(Role.Developer);

      await nextTick();

      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.false;
      expect(wrapper.emitted("update:name")).deep.equal([["My name"]]);
      expect(wrapper.emitted("update:role")).deep.equal([[Role.Developer]]);
    });

    it("should enable button when everything is valid and password is required", async () => {
      const wrapper = vuetifyMount(RoomForm, {
        props: {
          ...defaultProps,
          showPasswordInput: true,
        },
      });

      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;

      await wrapper.findAllComponents(VTextField).at(0).setValue("My name");
      await wrapper.findComponent(VRadioGroup).setValue(Role.Developer);

      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;
        
      await wrapper.findAllComponents(VTextField).at(1).setValue("top secret");

      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.false;
      expect(wrapper.emitted("update:password")).deep.equal([["top secret"]]);
    });

    it("should emit submit when form is submitted", async () => {
      const wrapper = vuetifyMount(RoomForm, {
        props: defaultProps,
      });
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;

      await wrapper.findAllComponents(VTextField).at(0).setValue("My name");
      await wrapper.findComponent(VRadioGroup).setValue(Role.Developer);

      await nextTick();

      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.false;
      await wrapper.findComponent(VBtn).trigger("click");
      expect(wrapper.emitted("submit")).deep.equal([[]]);
    });

    it("should show validation messages when fields are cleared", async () => {
      const wrapper = vuetifyMount(RoomForm, {
        props: defaultProps,
      });

      await wrapper.findAllComponents(VTextField).at(0).setValue("test 2");
      await wrapper.findAllComponents(VTextField).at(0).setValue("");
      await nextTick();
      expect(wrapper.findAllComponents(VTextField).at(0).text()).contains("Can not be empty");
    });

    it("should show validation message when text is too long", async () => {
      const wrapper = vuetifyMount(RoomForm, {
        props: defaultProps,
      });

      await wrapper.findAllComponents(VTextField).at(0).setValue("a".repeat(16));
      await nextTick();
      await nextTick();
      expect(wrapper.findAllComponents(VTextField).at(0).text()).contains("Only 15 chars allowed");
    });
  });
});
