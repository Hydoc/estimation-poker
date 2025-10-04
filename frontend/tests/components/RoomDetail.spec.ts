import { describe, expect, it, vi } from "vitest";
import RoomDetail from "../../src/components/RoomDetail.vue";
import { Role, RoundState } from "../../src/components/types";
import TableOverview from "../../src/components/TableOverview.vue";
import DeveloperCommandCenter from "../../src/components/DeveloperCommandCenter.vue";
import RoundSummary from "../../src/components/RoundSummary.vue";
import { nextTick } from "vue";
import { vuetifyMount } from "../vuetifyMount";

const ResizeObserverMock = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

vi.stubGlobal("ResizeObserver", ResizeObserverMock);
vi.stubGlobal("visualViewport", new EventTarget());
describe("RoomDetail", () => {
  describe("rendering", () => {
    it("should render", () => {
      const currentUsername = "Test";
      const usersInRoom = [
        { name: currentUsername, isDone: false, role: Role.Developer },
        { name: "Product Owner Test", role: Role.ProductOwner },
      ];
      const wrapper = vuetifyMount(RoomDetail, {
        props: {
          usersInRoom,
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "",
          guess: 0,
          didSkip: false,
          showAllGuesses: false,
          developerDone: [],
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      expect(wrapper.findComponent(TableOverview).exists()).to.be.true;
      expect(wrapper.findComponent(TableOverview).props("showAllGuesses")).to.be.false;
      expect(wrapper.findComponent(TableOverview).props("usersInRoom")).deep.equal([
        {
          isDone: false,
          name: "Test",
          role: "developer",
        },
        {
          name: "Product Owner Test",
          role: "product-owner",
        },
      ]);
      expect(wrapper.findComponent(TableOverview).props("ticketToGuess")).equal("");
      expect(wrapper.findComponent(TableOverview).props("roundState")).equal(RoundState.Waiting);

      expect(wrapper.findComponent(DeveloperCommandCenter).exists()).to.be.true;
      expect(wrapper.findComponent(DeveloperCommandCenter).props("showAllGuesses")).to.be.false;
      expect(wrapper.findComponent(DeveloperCommandCenter).props("guess")).equal(0);
      expect(wrapper.findComponent(DeveloperCommandCenter).props("didSkip")).to.be.false;
      expect(wrapper.findComponent(DeveloperCommandCenter).props("hasTicketToGuess")).to.be.false;
      expect(wrapper.findComponent(DeveloperCommandCenter).props("possibleGuesses")).deep.equal([
        { guess: 1, description: "Bis zu 4 Std." },
        { guess: 2, description: "Bis zu 8 Std." },
        { guess: 3, description: "Bis zu 3 Tagen" },
        { guess: 4, description: "Bis zu 5 Tagen" },
        { guess: 5, description: "Mehr als 5 Tage" },
      ]);
    });
  });

  describe("functionality", () => {
    it("should emit estimate when table overview emits estimate", () => {
      const wrapper = vuetifyMount(RoomDetail, {
        props: {
          developerDone: [],
          usersInRoom: [
            { name: "Test", guess: 0, role: Role.Developer },
            { name: "Product Owner Test", role: Role.ProductOwner },
          ],
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "",
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      wrapper.findComponent(TableOverview).vm.$emit("estimate", "WR-1");
      expect(wrapper.emitted("estimate")).deep.equal([["WR-1"]]);
    });

    it("should emit skip when developer command center emits skip", () => {
      const wrapper = vuetifyMount(RoomDetail, {
        props: {
          developerDone: [],
          usersInRoom: [
            { name: "Test", isDone: false, role: Role.Developer },
            { name: "Product Owner Test", role: Role.ProductOwner },
          ],
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "",
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      wrapper.findComponent(DeveloperCommandCenter).vm.$emit("skip", 1);
      expect(wrapper.emitted("skip")).deep.equal([[]]);
    });

    it("should emit guess when developer command center emits guess", () => {
      const wrapper = vuetifyMount(RoomDetail, {
        props: {
          developerDone: [],
          usersInRoom: [
            { name: "Test", isDone: false, role: Role.Developer },
            { name: "Product Owner Test", role: Role.ProductOwner },
          ],
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "",
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      wrapper.findComponent(DeveloperCommandCenter).vm.$emit("guess", 1);
      expect(wrapper.emitted("guess")).deep.equal([[1]]);
    });

    it("should emit reveal when table overview emits reveal", () => {
      const wrapper = vuetifyMount(RoomDetail, {
        props: {
          developerDone: [],
          usersInRoom: [
            { name: "Test", guess: 0, role: Role.Developer },
            { name: "Product Owner Test", role: Role.ProductOwner },
          ],
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "CC-1",
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      wrapper.findComponent(TableOverview).vm.$emit("reveal");
      expect(wrapper.emitted("reveal")).deep.equal([[]]);
    });

    it("should emit new round when table overview emits new round", () => {
      const wrapper = vuetifyMount(RoomDetail, {
        props: {
          developerDone: [],
          usersInRoom: [
            { name: "Test", guess: 0, role: Role.Developer },
            { name: "Product Owner Test", role: Role.ProductOwner },
          ],
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "CC-1",
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      wrapper.findComponent(TableOverview).vm.$emit("new-round");
      expect(wrapper.emitted("new-round")).deep.equal([[]]);
    });

    it("should show round summary depending if showAllGuesses is true", async () => {
      vi.useFakeTimers();
      const wrapper = vuetifyMount(RoomDetail, {
        props: {
          developerDone: [],
          usersInRoom: [
            { name: "Test", guess: 0, role: Role.Developer },
            { name: "Product Owner Test", role: Role.ProductOwner },
          ],
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "CC-1",
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      // @ts-ignore
      expect(wrapper.vm.showRoundSummary).to.be.false;
      expect(wrapper.findComponent(RoundSummary).exists()).to.be.false;

      await wrapper.setProps({
        showAllGuesses: true,
      });

      vi.runAllTimers();
      await nextTick();

      // @ts-ignore
      expect(wrapper.vm.showRoundSummary).to.be.true;
      expect(wrapper.findComponent(RoundSummary).exists()).to.be.true;

      await wrapper.setProps({
        showAllGuesses: false,
      });

      vi.runAllTimers();
      await nextTick();

      // @ts-ignore
      expect(wrapper.vm.showRoundSummary).to.be.false;
      expect(wrapper.findComponent(RoundSummary).exists()).to.be.false;
    });
  });
});
