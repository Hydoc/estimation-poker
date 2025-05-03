import { beforeEach, describe, expect, it, vi } from "vitest";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { mount } from "@vue/test-utils";
import RoundSummary from "../../src/components/RoundSummary.vue";
import { DeveloperDone } from "../../src/components/types";
import { VBottomSheet, VCard, VProgressCircular } from "vuetify/components";

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
      expect(wrapper.getComponent(VCard).text()).contains("2 Schätzungen");
      expect(wrapper.getComponent(VCard).text()).contains("Übereinstimmung");
      expect(wrapper.getComponent(VCard).getComponent(VProgressCircular).props("modelValue")).equal(
        100,
      );
    });

    it("should render for multiple cards including the coffee", () => {
      const wrapper = createWrapper([
        { guess: 1, role: "developer", name: "Test Dev 1", doSkip: false },
        { guess: 2, role: "developer", name: "Test Dev 2", doSkip: false },
        { guess: 0, role: "developer", name: "Test Dev 2", doSkip: true },
      ]);

      expect(wrapper.findComponent(VCard).exists()).to.be.true;
      expect(wrapper.getComponent(VCard).findAll(".card")).length(3);
      expect(wrapper.getComponent(VCard).findAll(".card").at(0).text()).equal("");
      expect(wrapper.getComponent(VCard).findAll(".card").at(1).text()).equal("1");
      expect(wrapper.getComponent(VCard).findAll(".card").at(2).text()).equal("2");

      expect(wrapper.getComponent(VCard).text()).contains("1 Schätzung");
      expect(wrapper.getComponent(VCard).text()).contains("Übereinstimmung");
      expect(wrapper.getComponent(VCard).getComponent(VProgressCircular).props("modelValue")).equal(
        33.33333333333333,
      );
    });
  });
});

function createWrapper(
  developerDone: DeveloperDone[] = [
    { guess: 2, role: "developer", name: "Test Dev 1", doSkip: false },
    { guess: 2, role: "developer", name: "Test Dev 2", doSkip: false },
  ],
) {
  return mount(RoundSummary, {
    props: {
      developerDone,
    },
    global: {
      plugins: [vuetify],
    },
  });
}
