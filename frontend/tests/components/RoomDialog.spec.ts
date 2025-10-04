import { describe, it, expect, vi } from "vitest";
import { vuetifyMount } from "../vuetifyMount";
import RoomDialog from "../../src/components/RoomDialog.vue";
import { Role } from "../../src/components/types";
import { VBtn, VDialog } from "vuetify/components";
import RoomForm from "../../src/components/RoomForm.vue";

vi.stubGlobal("visualViewport", new EventTarget());
describe("RoomDialog", () => {
  describe("rendering", () => {
    it("should render in closed state", () => {
      const wrapper = vuetifyMount(RoomDialog, {
        props: {
          activatorText: "Activator",
          cardTitle: "Title",
          name: "",
          role: Role.Empty,
          showDialog: false,
        },
      });

      expect(wrapper.findComponent(VDialog).exists()).to.be.true;
      expect(wrapper.findComponent(VDialog).props("modelValue")).to.be.false;
      expect(wrapper.findComponent(VDialog).props("width")).equal("500");
      expect(wrapper.findComponent(VBtn).exists()).to.be.true;
      expect(wrapper.findComponent(VBtn).props("color")).equal("primary");
      expect(wrapper.findComponent(VBtn).props("text")).equal("Activator");
      expect(wrapper.findComponent(RoomForm).exists()).to.be.false;
    });

    it("should render in open state with password input", () => {
      const wrapper = vuetifyMount(RoomDialog, {
        props: {
          activatorText: "Activator",
          cardTitle: "Title",
          name: "",
          role: Role.Empty,
          showDialog: true,
          showPasswordInput: true,
        },
      });

      expect(wrapper.findComponent(VDialog).exists()).to.be.true;
      expect(wrapper.findComponent(VDialog).props("modelValue")).to.be.true;
      expect(wrapper.findComponent(VDialog).props("width")).equal("500");
      expect(wrapper.findComponent(VBtn).exists()).to.be.true;
      expect(wrapper.findComponent(VBtn).props("color")).equal("primary");
      expect(wrapper.findComponent(VBtn).props("text")).equal("Activator");
      expect(wrapper.findComponent(RoomForm).exists()).to.be.true;
      expect(wrapper.findComponent(RoomForm).props("name")).equal("");
      expect(wrapper.findComponent(RoomForm).props("role")).equal(Role.Empty);
      expect(wrapper.findComponent(RoomForm).props("password")).equal("");
      expect(wrapper.findComponent(RoomForm).props("errorMessage")).to.be.undefined;
      expect(wrapper.findComponent(RoomForm).props("showPasswordInput")).to.be.true;
      expect(wrapper.findComponent(RoomForm).props("title")).equal("Title");
    });
  });

  describe("functionality", () => {
    it("should open dialog when activator was clicked", async () => {
      const wrapper = vuetifyMount(RoomDialog, {
        props: {
          activatorText: "Activator",
          cardTitle: "Title",
          name: "",
          role: Role.Empty,
          showDialog: false,
        },
      });

      await wrapper.findComponent(VBtn).trigger("click");

      expect(wrapper.findComponent(VDialog).props("modelValue")).to.be.true;
    });

    it("should emit v-model changes on room form", async () => {
      const wrapper = vuetifyMount(RoomDialog, {
        props: {
          activatorText: "Activator",
          cardTitle: "Title",
          name: "",
          role: Role.Empty,
          showDialog: true,
          showPasswordInput: true,
        },
      });

      wrapper.findComponent(RoomForm).vm.$emit("update:name", "New Name");

      expect(wrapper.emitted("update:name")).deep.equal([["New Name"]]);

      wrapper.findComponent(RoomForm).vm.$emit("update:role", Role.Developer);

      expect(wrapper.emitted("update:role")).deep.equal([[Role.Developer]]);

      wrapper.findComponent(RoomForm).vm.$emit("update:password", "New Password");

      expect(wrapper.emitted("update:password")).deep.equal([["New Password"]]);
    });

    it("should emit submit when room form emits it", () => {
      const wrapper = vuetifyMount(RoomDialog, {
        props: {
          activatorText: "Activator",
          cardTitle: "Title",
          name: "",
          role: Role.Empty,
          showDialog: true,
          showPasswordInput: true,
        },
      });

      wrapper.findComponent(RoomForm).vm.$emit("submit");

      expect(wrapper.emitted("submit")).deep.equal([[]]);
    });

    it("should pass error message to room form", () => {
      const wrapper = vuetifyMount(RoomDialog, {
        props: {
          activatorText: "Activator",
          cardTitle: "Title",
          name: "",
          role: Role.Empty,
          showDialog: true,
          showPasswordInput: true,
          errorMessage: "An error occurred.",
        },
      });

      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal("An error occurred.");
    });
  });
});
