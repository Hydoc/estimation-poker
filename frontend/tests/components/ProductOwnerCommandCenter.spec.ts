import { beforeEach, describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import ProductOwnerCommandCenter from "../../src/components/ProductOwnerCommandCenter.vue";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { VBtn, VForm, VTextField } from "vuetify/components";
import { nextTick } from "vue";

let vuetify: ReturnType<typeof createVuetify>;

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});
describe("ProductOwnerCommandCenter", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundIsWaiting: true,
          hasTicketToGuess: false,
          hasDevelopersInRoom: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VForm).exists()).to.be.true;
      expect(wrapper.findComponent(VForm).props("fastFail")).to.be.true;
      expect(wrapper.findComponent(VTextField).exists()).to.be.true;
      expect(wrapper.findComponent(VTextField).props("label")).equal("Ticket zum schätzen");
      expect(wrapper.findComponent(VTextField).props("placeholder")).equal("CC-0000");
      expect(Object.keys(wrapper.findComponent(VTextField).find("input").attributes())).contains(
        "required",
      );
      expect(wrapper.findComponent(VBtn).exists()).to.be.true;
      expect(wrapper.findComponent(VBtn).find("button").attributes("type")).equal("submit");
      expect(wrapper.findComponent(VBtn).text()).equal("Schätzen lassen");
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;
      expect(wrapper.find("p").exists()).to.be.false;
    });

    it("should render without developers in room", () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundIsWaiting: true,
          hasTicketToGuess: false,
          hasDevelopersInRoom: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VForm).exists()).to.be.false;
      expect(wrapper.find("p").text()).equal("Warten auf Entwickler...");
    });

    it("should not render form when round is finished", () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundIsWaiting: false,
          hasTicketToGuess: true,
          hasDevelopersInRoom: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VForm).exists()).to.be.false;
      expect(wrapper.find("div").text()).equal("");
    });
  });

  describe("functionality", () => {
    it("should enable button when everything is valid", async () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundIsWaiting: true,
          hasTicketToGuess: false,
          hasDevelopersInRoom: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findComponent(VTextField).setValue("WR-1");
      await nextTick();
      await nextTick();
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.false;
    });

    it("should show validation message when ticket is cleared", async () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundIsWaiting: true,
          hasTicketToGuess: false,
          hasDevelopersInRoom: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findComponent(VTextField).setValue("WR-1");
      await wrapper.findComponent(VTextField).setValue("");
      await nextTick();
      expect(wrapper.findComponent(VTextField).text()).contains(
        "Fehler: Hier müsste eigentlich was stehen",
      );
    });

    it("should show validation message when ticket does not match regex", async () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundIsWaiting: true,
          hasTicketToGuess: false,
          hasDevelopersInRoom: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findComponent(VTextField).setValue("WR");
      await wrapper.findComponent(VTextField).trigger("blur");
      await nextTick();
      expect(wrapper.findComponent(VTextField).text()).contains(
        "Fehler: Muss im Format ^[A-Z]{2}-\\d+$ sein",
      );
    });

    it("should emit estimate on form submit", async () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundIsWaiting: true,
          hasTicketToGuess: false,
          hasDevelopersInRoom: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findComponent(VTextField).setValue("WR-1");
      await nextTick();
      await nextTick();
      await wrapper.findComponent(VBtn).trigger("submit");

      expect(wrapper.emitted("estimate")).deep.equal([["WR-1"]]);
    });

    it("should not emit estimate when product owner can not estimate due to form invalid", async () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundIsWaiting: true,
          hasTicketToGuess: false,
          hasDevelopersInRoom: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      wrapper.vm.doLetEstimate();
      expect(wrapper.emitted("estimate")).to.be.undefined;
    });
  });
});
