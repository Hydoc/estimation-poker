import { beforeEach, describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import ProductOwnerCommandCenter from "../../src/components/ProductOwnerCommandCenter.vue";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import { VBtn, VForm, VProgressCircular, VTextField } from "vuetify/components";
import * as directives from "vuetify/directives";
import { nextTick } from "vue";
import { Developer, RoundState } from "../../src/components/types";

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
describe("ProductOwnerCommandCenter", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          showAllGuesses: false,
          developerList: [{ name: "Test", isDone: false, role: "developer" } as Developer],
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
      expect(wrapper.findAllComponents(VBtn)).length(1);
      expect(wrapper.findComponent(VBtn).find("button").attributes("type")).equal("submit");
      expect(wrapper.findComponent(VBtn).text()).equal("Schätzen lassen");
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;
      expect(wrapper.text()).not.contains("Warten auf Entwickler...");
    });

    it("should render without developers in room", () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          showAllGuesses: false,
          developerList: [],
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VForm).exists()).to.be.false;
      expect(wrapper.find("p").text()).equal("Warten auf Entwickler...");
    });

    it("should render button with progress bar when round is in progress but not every dev is done", () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.InProgress,
          hasTicketToGuess: true,
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: true, role: "developer" } as Developer,
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VForm).exists()).to.be.false;

      expect(wrapper.findAllComponents(VBtn)).length(1);
      expect(wrapper.findComponent(VBtn).text()).equal("Auflösen");
      expect(wrapper.findComponent(VBtn).props("loading")).to.be.true;
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;
      expect(wrapper.findComponent(VBtn).props("color")).equal("blue-grey");
      expect(wrapper.findComponent(VBtn).props("width")).equal("100%");

      expect(wrapper.findComponent(VProgressCircular).exists()).to.be.true;
      expect(wrapper.findComponent(VProgressCircular).props("modelValue")).equal(50);
    });

    it("should render button without progress bar when round is finished", () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.End,
          hasTicketToGuess: true,
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: true, role: "developer" } as Developer,
            { name: "Test 2", isDone: true, role: "developer" } as Developer,
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VForm).exists()).to.be.false;

      expect(wrapper.findAllComponents(VBtn)).length(1);
      expect(wrapper.findComponent(VBtn).text()).equal("Auflösen");
      expect(wrapper.findComponent(VBtn).props("loading")).to.be.false;
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.false;

      expect(wrapper.findComponent(VProgressCircular).exists()).to.be.false;
    });
  });

  describe("functionality", () => {
    it("should enable button when everything is valid", async () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: false, role: "developer" } as Developer,
          ],
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
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: false, role: "developer" } as Developer,
          ],
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
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: false, role: "developer" } as Developer,
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findComponent(VTextField).setValue("WR");
      await wrapper.findComponent(VTextField).trigger("blur");
      await nextTick();
      expect(wrapper.findComponent(VTextField).text()).contains(
        "Fehler: Muss im Format ^[A-Z]{2,}-\\d+$ sein",
      );
    });

    it("should emit estimate on form submit", async () => {
      const wrapper = mount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: false, role: "developer" } as Developer,
          ],
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
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: false, role: "developer" } as Developer,
          ],
        },
        global: {
          plugins: [vuetify],
        },
      });

      // @ts-ignore
      wrapper.vm.doLetEstimate();
      expect(wrapper.emitted("estimate")).to.be.undefined;
    });
  });
});
