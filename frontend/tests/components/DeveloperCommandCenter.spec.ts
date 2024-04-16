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

      expect(wrapper.findAll(".card")).length(5);
      expect(wrapper.findAll(".card").at(0).find("h2").text()).equal("1");
      expect(wrapper.findAll(".card").at(0).find("span").text()).equal("Bis zu 4 Std.");


      expect(wrapper.findAll(".card").at(1).find("h2").text()).equal("2");
      expect(wrapper.findAll(".card").at(1).find("span").text()).equal("Bis zu 8 Std.");

      expect(wrapper.findAll(".card").at(2).find("h2").text()).equal("3");
      expect(wrapper.findAll(".card").at(2).find("span").text()).equal("Bis zu 3 Tagen");

      expect(wrapper.findAll(".card").at(3).find("h2").text()).equal("4");
      expect(wrapper.findAll(".card").at(3).find("span").text()).equal("Bis zu 5 Tagen");

      expect(wrapper.findAll(".card").at(4).find("h2").text()).equal("5");
      expect(wrapper.findAll(".card").at(4).find("span").text()).equal("Mehr als 5 Tage");
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
    it("should emit guess on card click submit", async () => {
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

      await wrapper.findAll(".card").at(2).trigger("click");

      expect(wrapper.emitted("guess")).toEqual([[3]]);
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
  });
});
