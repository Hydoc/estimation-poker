import { beforeEach, describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import CommandCenter from "../../src/components/CommandCenter.vue";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { Role, RoundState } from "../../src/components/types";
import DeveloperCommandCenter from "../../src/components/DeveloperCommandCenter.vue";
import ProductOwnerCommandCenter from "../../src/components/ProductOwnerCommandCenter.vue";

let vuetify: ReturnType<typeof createVuetify>;

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});
describe("CommandCenter", () => {
  describe("rendering", () => {
    it("should render for developer", () => {
      const wrapper = mount(CommandCenter, {
        props: {
          showAllGuesses: false,
          didSkip: false,
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          guess: 0,
          ticketToGuess: "",
          hasDevelopersInRoom: true,
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

      expect(wrapper.findComponent(DeveloperCommandCenter).exists()).to.be.true;
      expect(wrapper.findComponent(ProductOwnerCommandCenter).exists()).to.be.false;

      expect(wrapper.findComponent(DeveloperCommandCenter).props("guess")).equal(0);
      expect(wrapper.findComponent(DeveloperCommandCenter).props("didSkip")).to.be.false;
      expect(wrapper.findComponent(DeveloperCommandCenter).props("hasTicketToGuess")).to.be.false;
    });

    it("should render for product owner", () => {
      const wrapper = mount(CommandCenter, {
        props: {
          userRole: Role.ProductOwner,
          didSkip: false,
          showAllGuesses: false,
          roundState: RoundState.Waiting,
          guess: 0,
          ticketToGuess: "",
          hasDevelopersInRoom: true,
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

      expect(wrapper.findComponent(ProductOwnerCommandCenter).exists()).to.be.true;
      expect(wrapper.findComponent(DeveloperCommandCenter).exists()).to.be.false;

      expect(wrapper.findComponent(ProductOwnerCommandCenter).props("roundIsWaiting")).to.be.true;
      expect(wrapper.findComponent(ProductOwnerCommandCenter).props("hasDevelopersInRoom")).to.be
        .true;
      expect(wrapper.findComponent(ProductOwnerCommandCenter).props("hasTicketToGuess")).to.be
        .false;
    });
  });

  describe("functionality", () => {
    it("should emit guess when developer guesses", () => {
      const wrapper = mount(CommandCenter, {
        props: {
          userRole: Role.Developer,
          didSkip: false,
          showAllGuesses: false,
          roundState: RoundState.Waiting,
          guess: 0,
          ticketToGuess: "",
          hasDevelopersInRoom: true,
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

      wrapper.findComponent(DeveloperCommandCenter).vm.$emit("guess", 3);
      expect(wrapper.emitted("guess")).deep.equal([[3]]);
    });

    it("should emit estimate when product owner estimates", () => {
      const wrapper = mount(CommandCenter, {
        props: {
          userRole: Role.ProductOwner,
          didSkip: false,
          showAllGuesses: false,
          roundState: RoundState.Waiting,
          guess: 0,
          ticketToGuess: "",
          hasDevelopersInRoom: true,
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

      wrapper.findComponent(ProductOwnerCommandCenter).vm.$emit("estimate", "WR-1");
      expect(wrapper.emitted("estimate")).deep.equal([["WR-1"]]);
    });
  });
});
