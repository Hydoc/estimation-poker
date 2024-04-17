import { beforeEach, describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import RoomDetail from "../../src/components/RoomDetail.vue";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { Role, RoundState } from "../../src/components/types";
import {
  VBtn,
  VIcon,
  VSnackbar,
  VDialog,
  VCard,
  VCardTitle,
  VTextField,
  VCardActions,
} from "vuetify/components";
import UserBox from "../../src/components/UserBox.vue";
import RoundOverview from "../../src/components/RoundOverview.vue";
import CommandCenter from "../../src/components/CommandCenter.vue";
import { nextTick } from "vue";

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
          didSkip: false,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: false,
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

      expect(wrapper.find("h1").text()).equal("öffentlicher Raum: ABC");
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

    it("should render when room is locked", () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: true,
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
      expect(wrapper.find("h1").text()).equal("privater Raum: ABC");
    });

    it("should render additional buttons when room is not locked and current user has permissions to lock", () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: true,
              key: "abc",
            },
          },
          roomIsLocked: false,
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

      expect(wrapper.find(".align-self-center").findAllComponents(VBtn)).length(2);
      expect(wrapper.find(".align-self-center").findAllComponents(VBtn).at(0).text()).equal(
        "Raum verlassen",
      );
      expect(wrapper.find(".align-self-center").findAllComponents(VBtn).at(1).text()).equal(
        "Raum schließen",
      );
      expect(
        wrapper.find(".align-self-center").findAllComponents(VBtn).at(1).props("appendIcon"),
      ).equal("mdi-lock");
      expect(wrapper.find(".align-self-center").findAllComponents(VBtn).at(1).props("color")).equal(
        "grey-darken-2",
      );
    });

    it("should render additional buttons when room is locked and current user has permissions to lock", () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: true,
              key: "abc",
            },
          },
          roomIsLocked: true,
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

      expect(wrapper.find(".align-self-center").findAllComponents(VBtn)).length(3);
      expect(wrapper.find(".align-self-center").findAllComponents(VBtn).at(0).text()).equal(
        "Raum verlassen",
      );
      expect(wrapper.find(".align-self-center").findAllComponents(VBtn).at(1).text()).equal(
        "Raum öffnen",
      );
      expect(
        wrapper.find(".align-self-center").findAllComponents(VBtn).at(1).props("appendIcon"),
      ).equal("mdi-key");
      expect(wrapper.find(".align-self-center").findAllComponents(VBtn).at(1).props("color")).equal(
        "grey-darken-2",
      );
      expect(wrapper.find(".align-self-center").findAllComponents(VBtn).at(2).text()).equal(
        "Passwort kopieren",
      );
      expect(
        wrapper.find(".align-self-center").findAllComponents(VBtn).at(2).props("appendIcon"),
      ).equal("mdi-content-copy");
      expect(wrapper.find(".align-self-center").findAllComponents(VBtn).at(2).props("color")).equal(
        "indigo-darken-3",
      );
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: false,
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: false,
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

      expect(wrapper.text()).not.contains("Raum verlassen");
    });
  });

  describe("functionality", () => {
    it("should copy room to clipboard when clicking mdi-content-copy and access is granted", async () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: false,
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

      await wrapper.findComponent(VIcon).trigger("click");
      expect(global.navigator.clipboard.writeText).toHaveBeenNthCalledWith(1, roomId);
      await nextTick();
      await nextTick();
      // @ts-ignore
      expect(wrapper.vm.snackbarText).equal("Kopiert!");
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: false,
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

      await wrapper.findComponent(VIcon).trigger("click");
      expect(global.navigator.clipboard.writeText).not.toHaveBeenCalled();
      await nextTick();
      await nextTick();
      // @ts-ignore
      expect(wrapper.vm.snackbarText).equal("Konnte nicht kopiert werden");
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: false,
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: false,
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: false,
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: false,
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: false,
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

      await wrapper.findComponent(VBtn).trigger("click");
      expect(wrapper.emitted("leave")).deep.equal([[]]);
    });

    it("should open dialog for setting password when 'Raum schließen' is clicked", async () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: true,
              key: "abc",
            },
          },
          roomIsLocked: false,
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

      await wrapper.find(".align-self-center").findAllComponents(VBtn).at(1).trigger("click");

      expect(wrapper.findComponent(VDialog).exists()).to.be.true;
      expect(wrapper.findComponent(VDialog).findComponent(VCard).exists()).to.be.true;
      expect(wrapper.findComponent(VDialog).findComponent(VCard).findComponent(VCardTitle).exists())
        .to.be.true;
      expect(
        wrapper.findComponent(VDialog).findComponent(VCard).findComponent(VCardTitle).text(),
      ).equal("Passwort setzen");
      expect(
        wrapper
          .findComponent(VDialog)
          .findComponent(VCard)
          .findComponent(VTextField)
          .props("placeholder"),
      ).equal("Passwort");
      expect(
        wrapper
          .findComponent(VDialog)
          .findComponent(VCard)
          .findComponent(VTextField)
          .props("appendIcon"),
      ).equal("mdi-eye-off");
      expect(
        wrapper.findComponent(VDialog).findComponent(VCard).findComponent(VTextField).props("type"),
      ).equal("password");

      expect(
        wrapper
          .findComponent(VDialog)
          .findComponent(VCard)
          .findComponent(VCardActions)
          .findAllComponents(VBtn),
      ).length(2);
      expect(
        wrapper
          .findComponent(VDialog)
          .findComponent(VCard)
          .findComponent(VCardActions)
          .findAllComponents(VBtn)
          .at(0)
          .text(),
      ).equal("Abbrechen");
      expect(
        wrapper
          .findComponent(VDialog)
          .findComponent(VCard)
          .findComponent(VCardActions)
          .findAllComponents(VBtn)
          .at(0)
          .props("color"),
      ).equal("red");
      expect(
        wrapper
          .findComponent(VDialog)
          .findComponent(VCard)
          .findComponent(VCardActions)
          .findAllComponents(VBtn)
          .at(1)
          .text(),
      ).equal("Abschließen");
      expect(
        wrapper
          .findComponent(VDialog)
          .findComponent(VCard)
          .findComponent(VCardActions)
          .findAllComponents(VBtn)
          .at(1)
          .props("color"),
      ).equal("green");
      expect(
        wrapper
          .findComponent(VDialog)
          .findComponent(VCard)
          .findComponent(VCardActions)
          .findAllComponents(VBtn)
          .at(1)
          .props("disabled"),
      ).to.be.true;
    });

    it("should toggle password field in dialog when eye icon is clicked", async () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: true,
              key: "abc",
            },
          },
          roomIsLocked: false,
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

      await wrapper.find(".align-self-center").findAllComponents(VBtn).at(1).trigger("click");

      // @ts-ignore
      wrapper.vm.showPassword = true;
      await nextTick();
      expect(
        wrapper
          .findComponent(VDialog)
          .findComponent(VCard)
          .findComponent(VTextField)
          .props("appendIcon"),
      ).equal("mdi-eye");
      expect(
        wrapper.findComponent(VDialog).findComponent(VCard).findComponent(VTextField).props("type"),
      ).equal("text");
    });

    it("should lock room when password was set", async () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: true,
              key: "abc",
            },
          },
          roomIsLocked: false,
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

      await wrapper.find(".align-self-center").findAllComponents(VBtn).at(1).trigger("click");
      await wrapper
        .findComponent(VDialog)
        .findComponent(VCard)
        .findComponent(VTextField)
        .setValue("top secret");

      expect(
        wrapper
          .findComponent(VDialog)
          .findComponent(VCard)
          .findComponent(VCardActions)
          .findAllComponents(VBtn)
          .at(1)
          .props("disabled"),
      ).to.be.false;

      await wrapper
        .findComponent(VDialog)
        .findComponent(VCard)
        .findComponent(VCardActions)
        .findAllComponents(VBtn)
        .at(1)
        .trigger("click");

      expect(wrapper.emitted("lock-room")).deep.equal([[{ key: "abc", password: "top secret" }]]);
    });

    it("should copy password when room is locked and user has permissions", async () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: true,
              key: "abc",
            },
          },
          roomIsLocked: true,
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

      // @ts-ignore
      wrapper.vm.roomPassword = "top secret";
      await wrapper.find(".align-self-center").findAllComponents(VBtn).at(2).trigger("click");
      await nextTick();
      // @ts-ignore
      expect(wrapper.vm.snackbarText).equal("Kopiert!");
      expect(global.navigator.clipboard.writeText).toHaveBeenNthCalledWith(1, "top secret");
    });

    it("should emit open-room when user with permission wants to open room", async () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: true,
              key: "abc",
            },
          },
          roomIsLocked: true,
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
      await wrapper.find(".align-self-center").findAllComponents(VBtn).at(1).trigger("click");
      expect(wrapper.emitted("open-room")).deep.equal([[{ key: "abc" }]]);
    });

    it("should use empty string as key when trying to open room without permission", () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: true,
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

      // @ts-ignore
      wrapper.vm.openRoom();
      expect(wrapper.emitted("open-room")).deep.equal([[{ key: "" }]]);
    });

    it("should use empty string as key when trying to lock room without permission", () => {
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
          didSkip: false,
          guess: 0,
          showAllGuesses: false,
          permissions: {
            room: {
              canLock: false,
            },
          },
          roomIsLocked: true,
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

      // @ts-ignore
      wrapper.vm.roomPassword = "top secret";
      // @ts-ignore
      wrapper.vm.lockRoom();
      expect(wrapper.emitted("lock-room")).deep.equal([[{ key: "", password: "top secret" }]]);
    });
  });
});
