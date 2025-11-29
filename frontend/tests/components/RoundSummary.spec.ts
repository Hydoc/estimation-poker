import { describe, expect, it, vi } from "vitest";
import { VBottomSheet, VCard, VProgressCircular } from "vuetify/components";
import RoundSummary from "../../src/components/RoundSummary.vue";
import { DeveloperDone, Role } from "../../src/components/types";
import { vuetifyMount } from "../vuetifyMount";

const ResizeObserverMock = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

vi.stubGlobal("ResizeObserver", ResizeObserverMock);
vi.stubGlobal("visualViewport", new EventTarget());
describe("RoundSummary", () => {
  describe("rendering", () => {
    it("should render for one card with multiple votes", () => {
      const wrapper = createWrapper();

      expect(wrapper.findComponent(VBottomSheet).exists()).to.be.true;
      expect(wrapper.findComponent(VBottomSheet).props("inset")).to.be.true;
      expect(wrapper.findComponent(VBottomSheet).props("height")).equal("250");
      expect(wrapper.findComponent(VBottomSheet).props("scrim")).to.be.false;

      expect(wrapper.findComponent(VCard).exists()).to.be.true;
      expect(wrapper.getComponent(VCard).findAll(".card")).length(1);
      expect(wrapper.getComponent(VCard).find(".card").text()).equal("2");
      expect(wrapper.getComponent(VCard).text()).contains("2 guesses");
      expect(wrapper.getComponent(VCard).text()).contains("Agreement");
      expect(wrapper.getComponent(VCard).getComponent(VProgressCircular).props("modelValue")).equal(
        100,
      );
    });

    it("should render for multiple cards including the coffee", () => {
      const wrapper = createWrapper([
        { guess: 1, role: Role.Developer, name: "Test Dev 1", doSkip: false },
        { guess: 2, role: Role.Developer, name: "Test Dev 2", doSkip: false },
        { guess: 0, role: Role.Developer, name: "Test Dev 2", doSkip: true },
      ]);

      expect(wrapper.findComponent(VCard).exists()).to.be.true;
      expect(wrapper.getComponent(VCard).findAll(".card")).length(3);
      expect(wrapper.getComponent(VCard).findAll(".card").at(0).text()).equal("");
      expect(wrapper.getComponent(VCard).findAll(".card").at(1).text()).equal("1");
      expect(wrapper.getComponent(VCard).findAll(".card").at(2).text()).equal("2");

      expect(wrapper.getComponent(VCard).text()).contains("1 guess");
      expect(wrapper.getComponent(VCard).text()).contains("Agreement");
      expect(wrapper.getComponent(VCard).getComponent(VProgressCircular).props("modelValue")).equal(
        33.33333333333333,
      );
    });
  });
});

function createWrapper(
  developerDone: DeveloperDone[] = [
    { guess: 2, role: Role.Developer, name: "Test Dev 1", doSkip: false },
    { guess: 2, role: Role.Developer, name: "Test Dev 2", doSkip: false },
  ],
) {
  return vuetifyMount(RoundSummary, {
    props: {
      developerDone,
    },
  });
}
