import { beforeEach, describe, expect, it, vi } from "vitest";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { mount } from "@vue/test-utils";
import TableOverview from "../../src/components/TableOverview.vue";
import {
  Developer,
  DeveloperDone,
  ProductOwner,
  Role,
  RoundState,
} from "../../src/components/types";
import ProductOwnerCommandCenter from "../../src/components/ProductOwnerCommandCenter.vue";
import DeveloperCard from "../../src/components/DeveloperCard.vue";

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
describe("TableOverview", () => {
  describe("rendering", () => {
    it("should render as product owner", () => {
      const wrapper = createWrapper(
        {
          productOwnerList: [{ name: "Test PO", role: "product-owner" } as ProductOwner],
          developerList: [{ name: "Test Dev", isDone: false, role: "developer" } as Developer],
        },
        RoundState.Waiting,
        [] as DeveloperDone[],
        false,
        Role.ProductOwner,
        "",
      );

      expect(wrapper.find(".virtual-table").exists()).to.be.true;
      expect(wrapper.find(".table").exists()).to.be.true;

      expect(wrapper.findComponent(ProductOwnerCommandCenter).exists()).to.be.true;
      expect(wrapper.findComponent(ProductOwnerCommandCenter).props("roundState")).equal(
        RoundState.Waiting,
      );
      expect(wrapper.findComponent(ProductOwnerCommandCenter).props("developerList")).deep.equal([
        { name: "Test Dev", isDone: false, role: "developer" } as Developer,
      ]);
      expect(wrapper.findComponent(ProductOwnerCommandCenter).props("hasTicketToGuess")).to.be
        .false;
      expect(wrapper.findComponent(ProductOwnerCommandCenter).props("showAllGuesses")).to.be.false;

      expect(wrapper.text()).not.contains("Warten auf Ticket…");

      expect(wrapper.findAll(".seat")).length(2);
      expect(wrapper.findAll(".seat").at(0).attributes("style")).equal("left: 285px; top: 50px;");
      expect(wrapper.findAll(".seat").at(1).attributes("style")).equal(
        "left: 283.00000000000006px; top: 550px;",
      );

      expect(wrapper.findAllComponents(DeveloperCard)).length(1);
      expect(wrapper.findAllComponents(DeveloperCard).at(0).props("developer")).deep.equal({
        name: "Test Dev",
        isDone: false,
        role: "developer",
      } as Developer);
      expect(wrapper.findAllComponents(DeveloperCard).at(0).props("developerDone")).to.be.undefined;

      expect(wrapper.findAll(".seat > span")).length(1);
      expect(wrapper.findAll(".seat > span").at(0).text()).equal("Test PO");
    });

    it("should render as developer", () => {
      const wrapper = createWrapper();

      expect(wrapper.findComponent(ProductOwnerCommandCenter).exists()).to.be.false;
      expect(wrapper.text()).contains("Warten auf Ticket…");
      expect(wrapper.findAllComponents(DeveloperCard)).length(1);
    });
  });

  describe("functionality", () => {
    it("should emit estimate when product owner command center emits estimate", () => {
      const wrapper = createWrapper(
        {
          productOwnerList: [{ name: "Test PO", role: "product-owner" } as ProductOwner],
          developerList: [{ name: "Test Dev", isDone: false, role: "developer" } as Developer],
        },
        RoundState.Waiting,
        [] as DeveloperDone[],
        false,
        Role.ProductOwner,
        "",
      );

      wrapper.findComponent(ProductOwnerCommandCenter).vm.$emit("estimate", "WH-12");
      expect(wrapper.emitted("estimate")).deep.equal([["WH-12"]]);
    });

    it("should emit reveal when product owner command center emits reveal", () => {
      const wrapper = createWrapper(
        {
          productOwnerList: [{ name: "Test PO", role: "product-owner" } as ProductOwner],
          developerList: [{ name: "Test Dev", isDone: false, role: "developer" } as Developer],
        },
        RoundState.Waiting,
        [] as DeveloperDone[],
        false,
        Role.ProductOwner,
        "",
      );

      wrapper.findComponent(ProductOwnerCommandCenter).vm.$emit("reveal");
      expect(wrapper.emitted("reveal")).deep.equal([[]]);
    });

    it("should emit new-round when product owner command center emits new-round", () => {
      const wrapper = createWrapper(
        {
          productOwnerList: [{ name: "Test PO", role: "product-owner" } as ProductOwner],
          developerList: [{ name: "Test Dev", isDone: false, role: "developer" } as Developer],
        },
        RoundState.Waiting,
        [] as DeveloperDone[],
        false,
        Role.ProductOwner,
        "WH-2",
      );

      wrapper.findComponent(ProductOwnerCommandCenter).vm.$emit("new-round");
      expect(wrapper.emitted("new-round")).deep.equal([[]]);
    });
  });
});

function createWrapper(
  usersInRoom = {
    productOwnerList: [{ name: "Test PO", role: "product-owner" } as ProductOwner],
    developerList: [{ name: "Test Dev", isDone: false, role: "developer" } as Developer],
  },
  roundState = RoundState.Waiting,
  developerDone = [] as DeveloperDone[],
  showAllGuesses = false,
  role = Role.Developer,
  ticketToGuess = "",
) {
  return mount(TableOverview, {
    props: {
      usersInRoom,
      roundState,
      developerDone,
      showAllGuesses,
      userRole: role,
      ticketToGuess,
    },
    global: {
      plugins: [vuetify],
    },
  });
}
