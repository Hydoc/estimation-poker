import { beforeEach, describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import DeveloperCommandCenter from "../../src/components/DeveloperCommandCenter.vue";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { VBtn, VCard, VCardSubtitle, VCardTitle, VItem, VItemGroup } from "vuetify/components";

let vuetify: ReturnType<typeof createVuetify>;

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});
describe("DeveloperCommandCenter", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = mount(DeveloperCommandCenter, {
        props: {
          didGuess: false,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VItemGroup).exists()).to.be.true;
      expect(wrapper.findComponent(VItemGroup).props("selectedClass")).equal("bg-indigo-darken-2");
      expect(wrapper.findAllComponents(VItem)).length(5);
      expect(wrapper.findAllComponents(VItem).at(0).props("value")).equal(1);
      expect(wrapper.findAllComponents(VItem).at(1).props("value")).equal(2);
      expect(wrapper.findAllComponents(VItem).at(2).props("value")).equal(3);
      expect(wrapper.findAllComponents(VItem).at(3).props("value")).equal(4);
      expect(wrapper.findAllComponents(VItem).at(4).props("value")).equal(5);

      expect(wrapper.findAllComponents(VCard)).length(5);
      wrapper.findAllComponents(VCard).forEach((it) => {
        expect(it.classes()).contains("text-center");
        expect(it.props("variant")).equal("outlined");
        expect(it.props("height")).equal("300");
        expect(it.props("link")).to.be.true;
      });

      expect(wrapper.findAllComponents(VCard).at(0).findComponent(VCardTitle).text()).equal("1");
      expect(wrapper.findAllComponents(VCard).at(0).findComponent(VCardSubtitle).text()).equal(
        "Bis zu 4 Std.",
      );

      expect(wrapper.findAllComponents(VCard).at(1).findComponent(VCardTitle).text()).equal("2");
      expect(wrapper.findAllComponents(VCard).at(1).findComponent(VCardSubtitle).text()).equal(
        "Bis zu 8 Std.",
      );

      expect(wrapper.findAllComponents(VCard).at(2).findComponent(VCardTitle).text()).equal("3");
      expect(wrapper.findAllComponents(VCard).at(2).findComponent(VCardSubtitle).text()).equal(
        "Bis zu 3 Tagen",
      );

      expect(wrapper.findAllComponents(VCard).at(3).findComponent(VCardTitle).text()).equal("4");
      expect(wrapper.findAllComponents(VCard).at(3).findComponent(VCardSubtitle).text()).equal(
        "Bis zu 5 Tagen",
      );

      expect(wrapper.findAllComponents(VCard).at(4).findComponent(VCardTitle).text()).equal("5");
      expect(wrapper.findAllComponents(VCard).at(4).findComponent(VCardSubtitle).text()).equal(
        "Mehr als 5 Tage",
      );

      expect(wrapper.findAllComponents(VBtn)).length(2);
      expect(wrapper.findAllComponents(VBtn).at(1).props("width")).equal("100%");
      expect(wrapper.findAllComponents(VBtn).at(1).props("prependIcon")).equal("mdi-send");
      expect(wrapper.findAllComponents(VBtn).at(1).props("appendIcon")).equal("mdi-send");
      expect(wrapper.findAllComponents(VBtn).at(1).props("disabled")).to.be.true;
      expect(wrapper.findAllComponents(VBtn).at(1).text()).equal("Ab gehts");
      expect(wrapper.find("p").exists()).to.be.false;
    });

    it("should render without ticket to guess", () => {
      const wrapper = mount(DeveloperCommandCenter, {
        props: {
          didGuess: false,
          didSkip: false,
          hasTicketToGuess: false,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.find("p").text()).equal("Warten auf Ticket...");
    });

    it("should not render when developer did guess", () => {
      const wrapper = mount(DeveloperCommandCenter, {
        props: {
          didGuess: true,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.find("div").text()).equal("");
    });
  });

  describe("functionality", () => {
    it("should enable button when everything is valid", async () => {
      const wrapper = mount(DeveloperCommandCenter, {
        props: {
          didGuess: false,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findAllComponents(VCard).at(2).trigger("click");
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.false;
    });

    it("should emit guess on form submit", async () => {
      const wrapper = mount(DeveloperCommandCenter, {
        props: {
          didGuess: false,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findAllComponents(VCard).at(2).trigger("click");
      await wrapper.findAllComponents(VBtn).at(1).trigger("click");

      expect(wrapper.emitted("guess")).toEqual([[3]]);
      // @ts-ignore
      expect(wrapper.vm.chosenCard).to.be.null;
    });

    it("should emit skip on skip button press", async () => {
      const wrapper = mount(DeveloperCommandCenter, {
        props: {
          didGuess: false,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });
      
      await wrapper.findComponent(VBtn).trigger("click");
      expect(wrapper.emitted("skip")).deep.equal([[]]);
    });
    it("should not emit guess when chosen card is null", () => {
      const wrapper = mount(DeveloperCommandCenter, {
        props: {
          didGuess: false,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });

      // @ts-ignore
      wrapper.vm.guess();

      expect(wrapper.emitted("guess")).to.be.undefined;
    });
  });
});
