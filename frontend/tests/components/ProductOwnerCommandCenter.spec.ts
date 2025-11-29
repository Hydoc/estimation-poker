import { describe, expect, it, vi } from "vitest";
import ProductOwnerCommandCenter from "../../src/components/ProductOwnerCommandCenter.vue";
import { VBtn, VForm, VProgressCircular, VTextField } from "vuetify/components";
import { nextTick } from "vue";
import { Developer, RoundState } from "../../src/components/types";
import { vuetifyMount } from "../vuetifyMount";

const ResizeObserverMock = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

vi.stubGlobal("ResizeObserver", ResizeObserverMock);
describe("ProductOwnerCommandCenter", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = vuetifyMount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          actualTicketToGuess: "",
          showAllGuesses: false,
          developerList: [{ name: "Test", isDone: false, role: "developer" } as Developer],
        },
      });

      expect(wrapper.findComponent(VForm).exists()).to.be.true;
      expect(wrapper.findComponent(VForm).props("fastFail")).to.be.true;
      expect(wrapper.findComponent(VTextField).exists()).to.be.true;
      expect(wrapper.findComponent(VTextField).props("label")).equal("Ticket to guess");
      expect(wrapper.findComponent(VTextField).props("placeholder")).equal("CC-0000");
      expect(Object.keys(wrapper.findComponent(VTextField).find("input").attributes())).contains(
        "required",
      );
      expect(wrapper.findAllComponents(VBtn)).length(1);
      expect(wrapper.findComponent(VBtn).find("button").attributes("type")).equal("submit");
      expect(wrapper.findComponent(VBtn).text()).equal("Estimate");
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;
      expect(wrapper.text()).not.contains("Waiting for developers...");
    });

    it("should render without developers in room", () => {
      const wrapper = vuetifyMount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          actualTicketToGuess: "",
          showAllGuesses: false,
          developerList: [],
        },
      });

      expect(wrapper.findComponent(VForm).exists()).to.be.false;
      expect(wrapper.find("p").text()).equal("Waiting for developers...");
    });

    it("should render button with progress bar when round is in progress but not every dev is done", () => {
      const wrapper = vuetifyMount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.InProgress,
          hasTicketToGuess: true,
          actualTicketToGuess: "WH-2",
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: true, role: "developer" } as Developer,
          ],
        },
      });

      expect(wrapper.findComponent(VForm).exists()).to.be.false;

      expect(wrapper.findAllComponents(VBtn)).length(1);
      expect(wrapper.findComponent(VBtn).text()).equal("Reveal");
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.true;
      expect(wrapper.findComponent(VBtn).props("color")).equal("teal");

      expect(wrapper.findComponent(VProgressCircular).exists()).to.be.true;
      expect(wrapper.findComponent(VProgressCircular).props("modelValue")).equal(50);
    });

    it("should render button without progress bar when round is finished", () => {
      const wrapper = vuetifyMount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.End,
          hasTicketToGuess: true,
          actualTicketToGuess: "WH-2",
          showAllGuesses: true,
          developerList: [
            { name: "Test", isDone: true, role: "developer" } as Developer,
            { name: "Test 2", isDone: true, role: "developer" } as Developer,
          ],
        },
      });

      expect(wrapper.findComponent(VForm).exists()).to.be.false;

      expect(wrapper.findAllComponents(VBtn)).length(1);
      expect(wrapper.findComponent(VBtn).text()).equal("New round");
      expect(wrapper.findComponent(VProgressCircular).exists()).to.be.false;
    });
  });

  describe("functionality", () => {
    it("should enable button when everything is valid", async () => {
      const wrapper = vuetifyMount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          actualTicketToGuess: "",
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: false, role: "developer" } as Developer,
          ],
        },
      });

      await wrapper.findComponent(VTextField).setValue("WR-1");
      await nextTick();
      await nextTick();
      expect(wrapper.findComponent(VBtn).props("disabled")).to.be.false;
    });

    it("should show validation message when ticket is cleared", async () => {
      const wrapper = vuetifyMount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          actualTicketToGuess: "",
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: false, role: "developer" } as Developer,
          ],
        },
      });

      await wrapper.findComponent(VTextField).setValue("WR-1");
      await wrapper.findComponent(VTextField).setValue("");
      await nextTick();
      expect(wrapper.findComponent(VTextField).text()).contains("Error: Can not be empty");
    });

    it("should show validation message when ticket does not match regex", async () => {
      const wrapper = vuetifyMount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          actualTicketToGuess: "",
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: false, role: "developer" } as Developer,
          ],
        },
      });

      await wrapper.findComponent(VTextField).setValue("WR");
      await wrapper.findComponent(VTextField).trigger("blur");
      await nextTick();
      expect(wrapper.findComponent(VTextField).text()).contains("Error: ^[A-Z]{2,}-\\d+$ required");
    });

    it("should emit estimate on form submit", async () => {
      const wrapper = vuetifyMount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          actualTicketToGuess: "",
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: false, role: "developer" } as Developer,
          ],
        },
      });

      await wrapper.findComponent(VTextField).setValue("WR-1");
      await nextTick();
      await nextTick();
      await wrapper.findComponent(VBtn).trigger("submit");

      expect(wrapper.emitted("estimate")).deep.equal([["WR-1"]]);
    });

    it("should not emit estimate when product owner can not estimate due to form invalid", async () => {
      const wrapper = vuetifyMount(ProductOwnerCommandCenter, {
        props: {
          roundState: RoundState.Waiting,
          hasTicketToGuess: false,
          actualTicketToGuess: "",
          showAllGuesses: false,
          developerList: [
            { name: "Test", isDone: false, role: "developer" } as Developer,
            { name: "Test 2", isDone: false, role: "developer" } as Developer,
          ],
        },
      });

      // @ts-ignore
      wrapper.vm.doLetEstimate();
      expect(wrapper.emitted("estimate")).to.be.undefined;
    });
  });
});
