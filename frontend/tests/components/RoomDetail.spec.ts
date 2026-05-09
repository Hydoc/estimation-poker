import { describe, expect, it, vi } from "vitest";
import RoomDetail from "../../src/components/RoomDetail.vue";
import TableOverview from "../../src/components/TableOverview.vue";
import DeveloperRoundView from "../../src/components/DeveloperRoundView.vue";
import RoundSummary from "../../src/components/RoundSummary.vue";
import { nextTick } from "vue";
import { vuetifyMount } from "../vuetifyMount";
import { RoundState } from "../../src/types/room";
import { nothing } from "@kaumlaut/pure/maybe";
import { RoomStateBuilder } from "../builder/RoomStateBuilder";

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
      const wrapper = createWrapper();

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
      expect(wrapper.findComponent(TableOverview).props("issueToGuess")).deep.equal(nothing());
      expect(wrapper.findComponent(TableOverview).props("roundState")).equal(RoundState.Waiting);

      expect(wrapper.findComponent(DeveloperRoundView).exists()).to.be.true;
      expect(wrapper.findComponent(DeveloperRoundView).props("showAllGuesses")).to.be.false;
      expect(wrapper.findComponent(DeveloperRoundView).props("guess")).deep.equal(nothing());
      expect(wrapper.findComponent(DeveloperRoundView).props("didSkip")).to.be.false;
      expect(wrapper.findComponent(DeveloperRoundView).props("hasIssueToGuess")).to.be.false;
      expect(wrapper.findComponent(DeveloperRoundView).props("possibleGuesses")).deep.equal([
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
      const wrapper = createWrapper();

      wrapper.findComponent(TableOverview).vm.$emit("estimate", "WR-1");
      expect(wrapper.emitted("estimate")).deep.equal([["WR-1"]]);
    });

    it("should emit skip when developer round view emits skip", () => {
      const wrapper = createWrapper();

      wrapper.findComponent(DeveloperRoundView).vm.$emit("skip", 1);
      expect(wrapper.emitted("skip")).deep.equal([[]]);
    });

    it("should emit guess when developer round view emits guess", () => {
      const wrapper = createWrapper();

      wrapper.findComponent(DeveloperRoundView).vm.$emit("guess", 1);
      expect(wrapper.emitted("guess")).deep.equal([[1]]);
    });

    it("should emit reveal when table overview emits reveal", () => {
      const wrapper = createWrapper();

      wrapper.findComponent(TableOverview).vm.$emit("reveal");
      expect(wrapper.emitted("reveal")).deep.equal([[]]);
    });

    it("should emit new round when table overview emits new round", () => {
      const wrapper = createWrapper();

      wrapper.findComponent(TableOverview).vm.$emit("new-round");
      expect(wrapper.emitted("new-round")).deep.equal([[]]);
    });

    it("should show round summary depending if showAllGuesses is true", async () => {
      vi.useFakeTimers();
      const wrapper = createWrapper(RoomStateBuilder.init().withIssueToGuess("CC-1"));

      // @ts-ignore
      expect(wrapper.vm.showRoundSummary).to.be.false;
      expect(wrapper.findComponent(RoundSummary).exists()).to.be.false;

      await wrapper.setProps({
        roomState: RoomStateBuilder.init().withShowAllGuesses(true).build(),
      });

      vi.runAllTimers();
      await nextTick();

      // @ts-ignore
      expect(wrapper.vm.showRoundSummary).to.be.true;
      expect(wrapper.findComponent(RoundSummary).exists()).to.be.true;

      await wrapper.setProps({
        roomState: RoomStateBuilder.init().build(),
      });

      vi.runAllTimers();
      await nextTick();

      // @ts-ignore
      expect(wrapper.vm.showRoundSummary).to.be.false;
      expect(wrapper.findComponent(RoundSummary).exists()).to.be.false;
    });
  });
});

function createWrapper(roomStateBuilder: RoomStateBuilder = RoomStateBuilder.init()) {
  return vuetifyMount(RoomDetail, {
    props: {
      roomState: roomStateBuilder.build(),
    },
  });
}
