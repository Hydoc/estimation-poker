import { beforeEach, describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import RoundOverview from "../../src/components/RoundOverview.vue";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { Role } from "../../src/components/types";
import { VBtn, VCard, VCardActions, VProgressCircular, VCardTitle } from "vuetify/components";
import ResultTable from "../../src/components/ResultTable.vue";

let vuetify: ReturnType<typeof createVuetify>;
const ResizeObserverMock = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

vi.stubGlobal("ResizeObserver", ResizeObserverMock);
beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});
describe("RoundOverview", () => {
  describe("rendering", () => {
    it("should render", () => {
      const developerList = [{ name: "Test", guess: 0, role: Role.Developer }];
      const wrapper = mount(RoundOverview, {
        props: {
          ticketToGuess: "WR-123",
          showAllGuesses: false,
          developerList: developerList,
          roundIsFinished: false,
          userIsProductOwner: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VCard).exists()).to.be.true;
      expect(wrapper.findComponent(VCard).findComponent(VCardTitle).find("span").text()).equal(
        "Aktuelles Ticket zum schätzen: WR-123",
      );

      expect(wrapper.findComponent(ResultTable).exists()).to.be.true;
      expect(wrapper.findComponent(ResultTable).props("developerList")).deep.equal(developerList);
      expect(wrapper.findComponent(ResultTable).props("showAllGuesses")).equal(false);
      expect(wrapper.findComponent(ResultTable).props("roundIsFinished")).equal(false);

      expect(wrapper.findComponent(VCardActions).exists()).to.be.false;
    });

    it("should render VCardActions when round is over, user is ProductOwner but not show all guesses", () => {
      const wrapper = mount(RoundOverview, {
        props: {
          ticketToGuess: "WR-123",
          showAllGuesses: false,
          developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
          roundIsFinished: true,
          userIsProductOwner: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VCardActions).exists()).to.be.true;
      expect(wrapper.findComponent(VCardActions).findAllComponents(VBtn)).length(1);
      expect(wrapper.findComponent(VCardActions).findAllComponents(VBtn).at(0).text()).equal(
        "Auflösen",
      );
      expect(
        wrapper.findComponent(VCardActions).findAllComponents(VBtn).at(0).props("color"),
      ).equal("primary");
    });

    it("should render VCardActions when round is over, user is ProductOwner and show all guesses", () => {
      const wrapper = mount(RoundOverview, {
        props: {
          ticketToGuess: "WR-123",
          showAllGuesses: true,
          developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
          roundIsFinished: true,
          userIsProductOwner: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VCardActions).exists()).to.be.true;
      expect(wrapper.findComponent(VCardActions).findAllComponents(VBtn)).length(1);
      expect(wrapper.findComponent(VCardActions).findAllComponents(VBtn).at(0).text()).equal(
        "Neue Runde",
      );
      expect(
        wrapper.findComponent(VCardActions).findAllComponents(VBtn).at(0).props("color"),
      ).equal("blue-darken-4");
    });

    it("should not render VCardActions when round is finished but user is not product owner", () => {
      const wrapper = mount(RoundOverview, {
        props: {
          ticketToGuess: "WR-123",
          showAllGuesses: true,
          developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
          roundIsFinished: true,
          userIsProductOwner: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VCardActions).exists()).to.be.false;
    });
  });

  describe("functionality", () => {
    it("should calculate percentage correctly", () => {
      const wrapper = mount(RoundOverview, {
        props: {
          ticketToGuess: "WR-123",
          showAllGuesses: false,
          developerList: [
            { name: "Test", guess: 2, doSkip: false, role: Role.Developer },
            { name: "Test", guess: 0, doSkip: true, role: Role.Developer },
            { name: "Test", guess: 0, doSkip: false, role: Role.Developer },
            { name: "Test", guess: 1, doSkip: false, role: Role.Developer },
            { name: "Test", guess: 0, doSkip: false, role: Role.Developer },
          ],
          roundIsFinished: false,
          userIsProductOwner: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VProgressCircular).props("modelValue")).equal(60);
      expect(wrapper.findComponent(VProgressCircular).props("color")).equal("teal-darken-1");
      expect(wrapper.findComponent(VProgressCircular).props("width")).equal("5");
      expect(wrapper.findComponent(VProgressCircular).props("size")).equal("50");
      expect(wrapper.findComponent(VProgressCircular).text()).equal("60%");
    });
    it("should emit on click 'Auflösen'", async () => {
      const wrapper = mount(RoundOverview, {
        props: {
          ticketToGuess: "WR-123",
          showAllGuesses: false,
          developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
          roundIsFinished: true,
          userIsProductOwner: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findComponent(VBtn).trigger("click");

      expect(wrapper.emitted("reveal")).deep.equal([[]]);
    });

    it("should emit on click 'Neue Runde'", async () => {
      const wrapper = mount(RoundOverview, {
        props: {
          ticketToGuess: "WR-123",
          showAllGuesses: true,
          developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
          roundIsFinished: true,
          userIsProductOwner: true,
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findComponent(VBtn).trigger("click");

      expect(wrapper.emitted("new-round")).deep.equal([[]]);
    });
  });
});
