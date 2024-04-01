import { beforeEach, describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import TheActiveRoomOverview from "../../src/components/TheActiveRoomOverview.vue";
import { VBtn, VCard, VDialog, VSheet, VTable } from "vuetify/components";
import { createVuetify } from "vuetify";
import RoomForm from "../../src/components/RoomForm.vue";
import { Role } from "../../src/components/types";

let vuetify: ReturnType<typeof createVuetify>;

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});
describe("TheActiveRoomOverview", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = mount(TheActiveRoomOverview, {
        props: {
          activeRooms: ["Hello", "world"],
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VSheet).exists()).to.be.true;
      expect(wrapper.findComponent(VDialog).exists()).to.be.true;
      expect(wrapper.findComponent(VDialog).props("modelValue")).to.be.false;
      expect(wrapper.find("h2").text()).equal("Bereits erstellte Räume");
      expect(wrapper.findComponent(VTable).exists()).to.be.true;
      expect(wrapper.findComponent(VTable).find("thead").text()).equal("Raum");
      expect(wrapper.findComponent(VTable).find("tbody").findAll("tr")).length(2);
      expect(
        wrapper.findComponent(VTable).find("tbody").findAll("tr").at(0).findAll("td").at(0).text(),
      ).equal("Hello");
      expect(
        wrapper
          .findComponent(VTable)
          .find("tbody")
          .findAll("tr")
          .at(0)
          .findAll("td")
          .at(1)
          .findComponent(VBtn)
          .text(),
      ).equal("Beitreten");
      expect(
        wrapper
          .findComponent(VTable)
          .find("tbody")
          .findAll("tr")
          .at(0)
          .findAll("td")
          .at(1)
          .findComponent(VBtn)
          .props("appendIcon"),
      ).equal("mdi-location-enter");
      expect(
        wrapper.findComponent(VTable).find("tbody").findAll("tr").at(1).findAll("td").at(0).text(),
      ).equal("world");
      expect(
        wrapper
          .findComponent(VTable)
          .find("tbody")
          .findAll("tr")
          .at(1)
          .findAll("td")
          .at(1)
          .findComponent(VBtn)
          .text(),
      ).equal("Beitreten");
      expect(
        wrapper
          .findComponent(VTable)
          .find("tbody")
          .findAll("tr")
          .at(1)
          .findAll("td")
          .at(1)
          .findComponent(VBtn)
          .props("appendIcon"),
      ).equal("mdi-location-enter");
    });
  });

  describe("functionality", () => {
    it("should show dialog when clicking Beitreten", async () => {
      const wrapper = mount(TheActiveRoomOverview, {
        props: {
          activeRooms: ["Hello", "world"],
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper
        .findComponent(VTable)
        .find("tbody")
        .findAll("tr")
        .at(0)
        .findAll("td")
        .at(1)
        .findComponent(VBtn)
        .trigger("click");

      expect(wrapper.findComponent(VDialog).props("modelValue")).to.be.true;
      expect(wrapper.findComponent(VDialog).props("width")).equal("500");
      expect(wrapper.findComponent(VDialog).findComponent(VCard).props("title")).equal(
        "Raum beitreten",
      );
      expect(wrapper.findComponent(VDialog).findComponent(RoomForm).exists()).to.be.true;
      expect(wrapper.findComponent(VDialog).findComponent(RoomForm).props("isRoomIdDisabled")).to.be
        .true;
      expect(wrapper.findComponent(VDialog).findComponent(RoomForm).props("roomId")).equal("Hello");
    });

    it("should emit join when room form submits", async () => {
      const wrapper = mount(TheActiveRoomOverview, {
        props: {
          activeRooms: ["Hello", "world"],
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper
        .findComponent(VTable)
        .find("tbody")
        .findAll("tr")
        .at(0)
        .findAll("td")
        .at(1)
        .findComponent(VBtn)
        .trigger("click");

      wrapper.vm.name = "Test";
      wrapper.vm.role = Role.Developer;

      wrapper.findComponent(RoomForm).vm.$emit("submit");

      expect(wrapper.emitted("join")).deep.equal([["Hello", "Test", "developer"]]);
    });
  });
});
