import { beforeEach, describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import RoomDetail from "../../src/components/RoomDetail.vue";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { Role, RoundState } from "../../src/components/types";
import { VBtn, VIcon, VSnackbar } from "vuetify/components";
import UserBox from "../../src/components/UserBox.vue";
import RoundOverview from "../../src/components/RoundOverview.vue";
import CommandCenter from "../../src/components/CommandCenter.vue";
import { nextTick } from "vue";

let vuetify: ReturnType<typeof createVuetify>;

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});
describe("RoomDetail", () => {
  describe("rendering", () => {
    it("should render", () => {
      const productOwnerList = [{ name: "Product Owner Test", role: Role.ProductOwner }];
      const currentUsername = "Test";
      const developerList = [{ name: currentUsername, guess: 0, role: Role.Developer }];
      const wrapper = mount(RoomDetail, {
        props: {
          roomId: "ABC",
          usersInRoom: {
            developerList,
            productOwnerList,
          },
          currentUsername: currentUsername,
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "",
          guess: 0,
          showAllGuesses: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.find("h1").text()).equal("Raum: ABC");
      expect(wrapper.find("h1").findComponent(VIcon).exists()).to.be.true;
      expect(wrapper.find("h1").findComponent(VIcon).find("i").attributes("title")).equal(
        "Raum kopieren",
      );
      expect(wrapper.find("h1").findComponent(VIcon).props("size")).equal("x-small");
      expect(wrapper.find("h1").findComponent(VIcon).find("i").classes()).contains(
        "mdi-content-copy",
      );

      expect(wrapper.findComponent(VBtn).exists()).to.be.true;
      expect(wrapper.findComponent(VBtn).props("color")).equal("deep-purple-darken-1");
      expect(wrapper.findComponent(VBtn).props("appendIcon")).equal("mdi-location-exit");
      expect(wrapper.findComponent(VBtn).text()).equal("Raum verlassen");

      expect(wrapper.findAllComponents(UserBox)).length(2);
      expect(wrapper.findAllComponents(UserBox).at(0).props("title")).equal("Product Owner");
      expect(wrapper.findAllComponents(UserBox).at(0).props("userList")).deep.equal(
        productOwnerList,
      );
      expect(wrapper.findAllComponents(UserBox).at(0).props("currentUsername")).equal(
        currentUsername,
      );
      expect(wrapper.findAllComponents(UserBox).at(1).props("title")).equal("Entwickler");
      expect(wrapper.findAllComponents(UserBox).at(1).props("userList")).deep.equal(developerList);
      expect(wrapper.findAllComponents(UserBox).at(1).props("currentUsername")).equal(
        currentUsername,
      );

      expect(wrapper.findComponent(RoundOverview).exists()).to.be.false;

      expect(wrapper.findComponent(CommandCenter).exists()).to.be.true;
      expect(wrapper.findComponent(CommandCenter).props("userRole")).equal(Role.Developer);
      expect(wrapper.findComponent(CommandCenter).props("roundState")).equal(RoundState.Waiting);
      expect(wrapper.findComponent(CommandCenter).props("guess")).equal(0);
      expect(wrapper.findComponent(CommandCenter).props("ticketToGuess")).equal("");
      expect(wrapper.findComponent(CommandCenter).props("hasDevelopersInRoom")).to.be.true;
      expect(wrapper.findComponent(VSnackbar).exists()).to.be.true;
      expect(wrapper.findComponent(VSnackbar).props("modelValue")).to.be.false;
    });

    it("should render RoundOverview when ticketToGuess != ''", () => {
      const productOwnerList = [{ name: "Product Owner Test", role: Role.ProductOwner }];
      const currentUsername = "Test";
      const developerList = [{ name: currentUsername, guess: 0, role: Role.Developer }];
      const wrapper = mount(RoomDetail, {
        props: {
          roomId: "ABC",
          usersInRoom: {
            developerList,
            productOwnerList,
          },
          currentUsername: currentUsername,
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "CC-1",
          guess: 0,
          showAllGuesses: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(RoundOverview).exists()).to.be.true;
      expect(wrapper.findComponent(RoundOverview).props("roundIsFinished")).to.be.false;
      expect(wrapper.findComponent(RoundOverview).props("showAllGuesses")).to.be.false;
      expect(wrapper.findComponent(RoundOverview).props("developerList")).deep.equal(developerList);
      expect(wrapper.findComponent(RoundOverview).props("ticketToGuess")).equal("CC-1");
      expect(wrapper.findComponent(RoundOverview).props("userIsProductOwner")).to.be.false;
    });

    it("should not render leave room when round has begun", () => {
      const productOwnerList = [{ name: "Product Owner Test", role: Role.ProductOwner }];
      const currentUsername = "Test";
      const developerList = [{ name: currentUsername, guess: 0, role: Role.Developer }];
      const wrapper = mount(RoomDetail, {
        props: {
          roomId: "ABC",
          usersInRoom: {
            developerList,
            productOwnerList,
          },
          currentUsername: currentUsername,
          userRole: Role.Developer,
          roundState: RoundState.InProgress,
          ticketToGuess: "CC-1",
          guess: 0,
          showAllGuesses: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.text()).not.contains("Raum verlassen");
    });
  });

  describe("functionality", () => {
    it("should copy room to clipboard when clicking mdi-content-copy and access is granted", async () => {
      const originalNavigator = global.navigator;
      Object.defineProperty(global.navigator, "clipboard", {
        writable: true,
        value: {
          writeText: vi.fn(),
        },
      });
      Object.defineProperty(global.navigator, "permissions", {
        writable: true,
        value: {
          query: vi.fn().mockResolvedValue({ state: "granted" }),
        },
      });
      const roomId = "ABC";
      const wrapper = mount(RoomDetail, {
        props: {
          roomId,
          usersInRoom: {
            developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
            productOwnerList: [{ name: "Product Owner Test", role: Role.ProductOwner }],
          },
          currentUsername: "Test",
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "CC-1",
          guess: 0,
          showAllGuesses: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findComponent(VIcon).trigger("click");
      expect(global.navigator.clipboard.writeText).toHaveBeenNthCalledWith(1, roomId);
      await nextTick();
      await nextTick();
      // @ts-ignore
      expect(wrapper.vm.snackbarText).equal("Raum in die Zwischenablage kopiert!");
      expect(wrapper.findComponent(VSnackbar).exists()).to.be.true;
      expect(wrapper.findComponent(VSnackbar).props("modelValue")).to.be.true;
      expect(wrapper.findComponent(VSnackbar).props("timeout")).equal(3000);
      expect(global.navigator.permissions.query).toHaveBeenNthCalledWith(1, {
        name: "clipboard-write",
      });
    });

    it("should not copy room to clipboard when clicking mdi-content-copy and access is not granted", async () => {
      Object.defineProperty(global.navigator, "clipboard", {
        writable: true,
        value: {
          writeText: vi.fn(),
        },
      });
      Object.defineProperty(global.navigator, "permissions", {
        writable: true,
        value: {
          query: vi.fn().mockResolvedValue({ state: "denied" }),
        },
      });
      const roomId = "ABC";
      const wrapper = mount(RoomDetail, {
        props: {
          roomId,
          usersInRoom: {
            developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
            productOwnerList: [{ name: "Product Owner Test", role: Role.ProductOwner }],
          },
          currentUsername: "Test",
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "CC-1",
          guess: 0,
          showAllGuesses: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findComponent(VIcon).trigger("click");
      expect(global.navigator.clipboard.writeText).not.toHaveBeenCalled();
      await nextTick();
      await nextTick();
      // @ts-ignore
      expect(wrapper.vm.snackbarText).equal("Konnte Raum nicht in die Zwischenablage kopieren");
      expect(wrapper.findComponent(VSnackbar).exists()).to.be.true;
      expect(wrapper.findComponent(VSnackbar).props("modelValue")).to.be.true;
      expect(wrapper.findComponent(VSnackbar).props("timeout")).equal(3000);
      expect(global.navigator.permissions.query).toHaveBeenNthCalledWith(1, {
        name: "clipboard-write",
      });
    });

    it("should emit estimate when command center emits estimate", () => {
      const roomId = "ABC";
      const wrapper = mount(RoomDetail, {
        props: {
          roomId,
          usersInRoom: {
            developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
            productOwnerList: [{ name: "Product Owner Test", role: Role.ProductOwner }],
          },
          currentUsername: "Test",
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "",
          guess: 0,
          showAllGuesses: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      wrapper.findComponent(CommandCenter).vm.$emit("estimate", "WR-1");
      expect(wrapper.emitted("estimate")).deep.equal([["WR-1"]]);
    });

    it("should emit guess when command center emits guess", () => {
      const roomId = "ABC";
      const wrapper = mount(RoomDetail, {
        props: {
          roomId,
          usersInRoom: {
            developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
            productOwnerList: [{ name: "Product Owner Test", role: Role.ProductOwner }],
          },
          currentUsername: "Test",
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "",
          guess: 0,
          showAllGuesses: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      wrapper.findComponent(CommandCenter).vm.$emit("guess", 1);
      expect(wrapper.emitted("guess")).deep.equal([[1]]);
    });

    it("should emit reveal when round overview emits reveal", () => {
      const roomId = "ABC";
      const wrapper = mount(RoomDetail, {
        props: {
          roomId,
          usersInRoom: {
            developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
            productOwnerList: [{ name: "Product Owner Test", role: Role.ProductOwner }],
          },
          currentUsername: "Test",
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "CC-1",
          guess: 0,
          showAllGuesses: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      wrapper.findComponent(RoundOverview).vm.$emit("reveal");
      expect(wrapper.emitted("reveal")).deep.equal([[]]);
    });

    it("should emit new round when round overview emits new round", () => {
      const roomId = "ABC";
      const wrapper = mount(RoomDetail, {
        props: {
          roomId,
          usersInRoom: {
            developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
            productOwnerList: [{ name: "Product Owner Test", role: Role.ProductOwner }],
          },
          currentUsername: "Test",
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "CC-1",
          guess: 0,
          showAllGuesses: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      wrapper.findComponent(RoundOverview).vm.$emit("new-round");
      expect(wrapper.emitted("new-round")).deep.equal([[]]);
    });

    it("should emit leave when leave button was clicked", async () => {
      const roomId = "ABC";
      const wrapper = mount(RoomDetail, {
        props: {
          roomId,
          usersInRoom: {
            developerList: [{ name: "Test", guess: 0, role: Role.Developer }],
            productOwnerList: [{ name: "Product Owner Test", role: Role.ProductOwner }],
          },
          currentUsername: "Test",
          userRole: Role.Developer,
          roundState: RoundState.Waiting,
          ticketToGuess: "CC-1",
          guess: 0,
          showAllGuesses: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      await wrapper.findComponent(VBtn).trigger("click");
      expect(wrapper.emitted("leave")).deep.equal([[]]);
    });
  });
});
