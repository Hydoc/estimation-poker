import { beforeEach, describe, expect, it, Mock, vi } from "vitest";
import RoomView from "../../src/views/RoomView.vue";
import { createTestingPinia, TestingPinia } from "@pinia/testing";
import { useRoute, useRouter } from "vue-router";
import RoomDetail from "../../src/components/RoomDetail.vue";
import { VBtn, VDialog, VIcon, VTextField, VToolbar } from "vuetify/components";
import { nextTick } from "vue";
import { vuetifyMount } from "../vuetifyMount";
import RoomForm from "../../src/components/RoomForm.vue";
import { useEstimationStore } from "../../src/stores/estimation";
import { RoomStateBuilder } from "../builder/RoomStateBuilder";
import { Role, RoundState } from "../../src/types/room";

vi.mock("vue-router");

let pinia: TestingPinia;
let estimationStore: ReturnType<typeof useEstimationStore>;
const defaultRoomState = RoomStateBuilder.init()
  .withId("ABC")
  .withName("test dev")
  .withRole(Role.Developer)
  .withUsers([
    { name: "test po", role: Role.ProductOwner },
    { name: "test dev", role: Role.Developer, isDone: false },
  ])
  .withPermissions({
    room: {
      canLock: true,
      key: "abc",
    },
  });

const ResizeObserverMock = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));
vi.stubGlobal("ResizeObserver", ResizeObserverMock);
vi.stubGlobal("visualViewport", new EventTarget());
beforeEach(() => {
  (useRouter as Mock).mockReturnValue({
    push: vi.fn(),
  });

  (useRoute as Mock).mockReturnValue({
    params: {},
  });

  pinia = createTestingPinia();
  estimationStore = useEstimationStore(pinia);
  // @ts-ignore
  estimationStore.roomNotifications = [];
  // @ts-ignore
  estimationStore.roomState = defaultRoomState.build();
  // @ts-ignore
  estimationStore.roomMetadata = vi.fn(() =>
    Promise.resolve({
      exists: true,
      isLocked: false,
    }),
  );
});
describe("RoomView", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = createWrapper();

      expect(wrapper.findComponent(VToolbar).exists()).to.be.true;
      expect(wrapper.findComponent(VToolbar).props("rounded")).to.be.true;

      expect(wrapper.findComponent(RoomDetail).exists()).to.be.true;
      expect(wrapper.findComponent(RoomDetail).props("roomState")).deep.equal(
        // @ts-ignore
        estimationStore.roomState,
      );
    });
  });

  describe("functionality", () => {
    it("should render room form when user is not connected to the room", async () => {
      // @ts-ignore
      estimationStore.roomState = defaultRoomState.withConnected(false).build();
      const wrapper = createWrapper();

      expect(wrapper.findComponent(RoomDetail).exists()).to.be.false;
      expect(wrapper.findComponent(RoomForm).exists()).to.be.true;
      expect(wrapper.findComponent(RoomForm).props("title")).equal("Join room");
      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal("");
      expect(wrapper.findComponent(RoomForm).props("showPasswordInput")).to.be.false;
      expect(wrapper.findComponent(RoomForm).props("subtitle")).equal(
        "You are currently not connected to this room",
      );
    });

    it("should render room form when user is not connected to the room and password is required", async () => {
      // @ts-ignore
      estimationStore.roomState = defaultRoomState.withConnected(false).build();
      // @ts-ignore
      estimationStore.roomMetadata = vi.fn(() =>
        Promise.resolve({
          exists: true,
          isLocked: true,
        }),
      );
      const wrapper = createWrapper();

      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(RoomForm).exists()).to.be.true;
      expect(wrapper.findComponent(RoomForm).props("showPasswordInput")).to.be.true;
    });

    it("should send correct message on estimate", () => {
      const wrapper = createWrapper();

      wrapper.findComponent(RoomDetail).vm.$emit("estimate", "WR-12");
      // @ts-ignore
      expect(estimationStore.send).toHaveBeenNthCalledWith(1, {
        type: "estimate",
        data: "WR-12",
      });
    });

    it("should send correct message on guess", () => {
      const wrapper = createWrapper();

      wrapper.findComponent(RoomDetail).vm.$emit("guess", 2);
      // @ts-ignore
      expect(estimationStore.send).toHaveBeenNthCalledWith(1, {
        type: "guess",
        data: 2,
      });
    });

    it("should send correct message on reveal", () => {
      const wrapper = createWrapper();

      wrapper.findComponent(RoomDetail).vm.$emit("reveal");
      // @ts-ignore
      expect(estimationStore.send).toHaveBeenNthCalledWith(1, {
        type: "reveal",
        data: null,
      });
    });

    it("should redirect to '/' when room does not exists", async () => {
      // @ts-ignore
      estimationStore.roomState = defaultRoomState.withConnected(false).build();
      // @ts-ignore
      estimationStore.roomMetadata = vi.fn(() =>
        Promise.resolve({
          exists: false,
          isLocked: false,
        }),
      );
      createWrapper();

      await nextTick();

      expect(useRouter().push).toHaveBeenNthCalledWith(1, "/");
    });

    it("should send correct message on new-round", () => {
      const wrapper = createWrapper();

      wrapper.findComponent(RoomDetail).vm.$emit("new-round");
      // @ts-ignore
      expect(estimationStore.send).toHaveBeenNthCalledWith(1, {
        type: "new-round",
        data: null,
      });
    });

    it("should disconnect and push to home when on leave", () => {
      const wrapper = createWrapper();

      wrapper
        .findComponent(VToolbar)
        .findAllComponents(VBtn)
        .find((btn) => btn.findComponent(VIcon).find("i").classes().includes("mdi-location-exit"))
        .trigger("click");

      // @ts-ignore
      expect(estimationStore.leaveRoom).toHaveBeenCalledOnce();
      expect(useRouter().push).toHaveBeenNthCalledWith(1, "/");
    });

    it("should send correct message when locking room", async () => {
      const wrapper = createWrapper();

      await wrapper
        .findComponent(VToolbar)
        .findAllComponents(VBtn)
        .find((btn) => btn.findComponent(VIcon).find("i").classes().includes("mdi-lock"))
        .trigger("click");

      expect(wrapper.findComponent(VDialog).exists()).to.be.true;

      await wrapper.findComponent(VDialog).findComponent(VTextField).setValue("top secret");
      await wrapper
        .findComponent(VDialog)
        .findAllComponents(VBtn)
        .find((btn) => btn.text() === "Lock")
        .trigger("click");

      // @ts-ignore
      expect(estimationStore.send).toHaveBeenNthCalledWith(1, {
        type: "lock-room",
        data: { password: "top secret", key: "abc" },
      });
    });

    it("should send correct message when opening room", async () => {
      // @ts-ignore
      estimationStore.roomState = defaultRoomState.withRoomIsLocked(true).build();
      const wrapper = createWrapper();

      await wrapper
        .findComponent(VToolbar)
        .findAllComponents(VBtn)
        .find((btn) => btn.findComponent(VIcon).find("i").classes().includes("mdi-key"))
        .trigger("click");

      // @ts-ignore
      expect(estimationStore.send).toHaveBeenNthCalledWith(1, {
        type: "open-room",
        data: { key: "abc" },
      });
    });

    it("should copy password when room is locked and user has permissions", async () => {
      // @ts-ignore
      estimationStore.roomState = defaultRoomState
        .withRoomIsLocked(true)
        .withPermissions({
          room: {
            canLock: true,
            key: "",
          },
        })
        .build();
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

      const wrapper = createWrapper();

      // @ts-ignore
      wrapper.vm.roomPassword = "top secret";
      await wrapper
        .findComponent(VToolbar)
        .findAllComponents(VBtn)
        .find((btn) => btn.findComponent(VIcon).find("i").classes().includes("mdi-content-copy"))
        .trigger("click");
      await nextTick();
      // @ts-ignore
      expect(global.navigator.clipboard.writeText).toHaveBeenNthCalledWith(1, "top secret");
    });

    it("should not copy password when room is locked but no permission is granted", async () => {
      // @ts-ignore
      estimationStore.roomState = defaultRoomState.withRoomIsLocked(true).build();
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

      const wrapper = createWrapper();

      // @ts-ignore
      wrapper.vm.roomPassword = "top secret";
      await wrapper
        .findComponent(VToolbar)
        .findAllComponents(VBtn)
        .find((btn) => btn.findComponent(VIcon).find("i").classes().includes("mdi-content-copy"))
        .trigger("click");
      await nextTick();
      expect(global.navigator.clipboard.writeText).not.toHaveBeenCalled();
    });

    it("should join when user not connected and everything is fine", async () => {
      // @ts-ignore
      estimationStore.roomState = defaultRoomState.withConnected(false).build();
      // @ts-ignore
      estimationStore.roomMetadata = vi.fn(() =>
        Promise.resolve({
          exists: true,
          isLocked: false,
        }),
      );
      // @ts-ignore
      estimationStore.joinRoom = vi.fn();
      (useRoute as Mock).mockReturnValue({
        params: {
          id: "room-id",
        },
      });
      const wrapper = createWrapper();

      wrapper.findComponent(RoomForm).vm.$emit("update:name", "Name");
      wrapper.findComponent(RoomForm).vm.$emit("update:role", Role.Developer);
      wrapper.findComponent(RoomForm).vm.$emit("submit");

      await nextTick();
      await nextTick();
      await nextTick();

      // @ts-ignore
      expect(estimationStore.fetchRoomState).toHaveBeenCalledOnce();
      // @ts-ignore
      expect(estimationStore.joinRoom).toHaveBeenNthCalledWith(
        1,
        "Name",
        Role.Developer,
        "room-id",
      );
      // @ts-ignore
      expect(estimationStore.fetchPermissions).toHaveBeenCalledOnce();
    });

    it("should not join when user is not connected, room is locked and password does not match", async () => {
      // @ts-ignore
      estimationStore.roomState = defaultRoomState.withConnected(false).build();
      // @ts-ignore
      estimationStore.roomMetadata = vi.fn(() =>
        Promise.resolve({
          exists: true,
          isLocked: true,
        }),
      );

      // @ts-ignore
      estimationStore.connectionState = vi.fn(() =>
        Promise.resolve({
          canConnect: false,
          reason: "wrong password",
        }),
      );

      // @ts-ignore
      estimationStore.joinRoom = vi.fn();
      (useRoute as Mock).mockReturnValue({
        params: {
          id: "room-id",
        },
      });
      const wrapper = createWrapper();

      wrapper.findComponent(RoomForm).vm.$emit("update:name", "Name");
      wrapper.findComponent(RoomForm).vm.$emit("update:role", Role.Developer);
      wrapper.findComponent(RoomForm).vm.$emit("update:password", "incorrect");
      wrapper.findComponent(RoomForm).vm.$emit("submit");

      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal(
        "The provided password is wrong",
      );

      // @ts-ignore
      expect(estimationStore.fetchRoomState).not.toHaveBeenCalled();
      // @ts-ignore
      expect(estimationStore.joinRoom).not.toHaveBeenCalled();
      // @ts-ignore
      expect(estimationStore.fetchPermissions).not.toHaveBeenCalled();
    });

    it("should not join when user is not connected, password matches but round already started", async () => {
      // @ts-ignore
      estimationStore.roomState = defaultRoomState.withConnected(false).build();
      // @ts-ignore
      estimationStore.connectionState = vi.fn(() =>
        Promise.resolve({
          canConnect: false,
          reason: "round already started",
        }),
      );
      // @ts-ignore
      estimationStore.roomMetadata = vi.fn(() =>
        Promise.resolve({
          exists: true,
          isLocked: true,
        }),
      );
      // @ts-ignore
      estimationStore.joinRoom = vi.fn();
      (useRoute as Mock).mockReturnValue({
        params: {
          id: "room-id",
        },
      });
      const wrapper = createWrapper();

      wrapper.findComponent(RoomForm).vm.$emit("update:name", "Name");
      wrapper.findComponent(RoomForm).vm.$emit("update:role", Role.Developer);
      wrapper.findComponent(RoomForm).vm.$emit("update:password", "correct");
      wrapper.findComponent(RoomForm).vm.$emit("submit");

      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal(
        "The round has already started",
      );

      // @ts-ignore
      expect(estimationStore.fetchRoomState).not.toHaveBeenCalled();
      // @ts-ignore
      expect(estimationStore.joinRoom).not.toHaveBeenCalled();
      // @ts-ignore
      expect(estimationStore.fetchPermissions).not.toHaveBeenCalled();
    });

    it("should not join when user is not connected, password matches but user already exists in the room", async () => {
      // @ts-ignore
      estimationStore.roomState = defaultRoomState.withConnected(false).build();
      // @ts-ignore
      estimationStore.roomMetadata = vi.fn(() =>
        Promise.resolve({
          exists: true,
          isLocked: true,
        }),
      );
      // @ts-ignore
      estimationStore.connectionState = vi.fn(() =>
        Promise.resolve({
          canConnect: false,
          reason: "username already taken",
        }),
      );
      // @ts-ignore
      estimationStore.joinRoom = vi.fn();
      (useRoute as Mock).mockReturnValue({
        params: {
          id: "room-id",
        },
      });
      const wrapper = createWrapper();

      wrapper.findComponent(RoomForm).vm.$emit("update:name", "Name");
      wrapper.findComponent(RoomForm).vm.$emit("update:role", Role.Developer);
      wrapper.findComponent(RoomForm).vm.$emit("update:password", "correct");
      wrapper.findComponent(RoomForm).vm.$emit("submit");

      await nextTick();
      await nextTick();
      await nextTick();

      expect(wrapper.findComponent(RoomForm).props("errorMessage")).equal(
        "A user with this name already exists in the room",
      );

      // @ts-ignore
      expect(estimationStore.fetchRoomState).not.toHaveBeenCalled();
      // @ts-ignore
      expect(estimationStore.joinRoom).not.toHaveBeenCalled();
      // @ts-ignore
      expect(estimationStore.fetchPermissions).not.toHaveBeenCalled();
    });
  });

  it("should return first id in params when array is given", () => {
    (useRoute as Mock).mockReturnValue({
      params: {
        id: ["first", "second"],
      },
    });

    const wrapper = createWrapper();

    // @ts-ignore
    expect(wrapper.vm.queryRoomId).equal("first");
  });

  it("should return correct roundStateAsReadableString when round is in Progress an there is an issue to guess", () => {
    // @ts-ignore
    estimationStore.roomState = defaultRoomState
      .withRoundState(RoundState.InProgress)
      .withIssueToGuess("AC-2")
      .build();
    const wrapper = createWrapper();

    // @ts-ignore
    expect(wrapper.vm.roundStateAsReadableString).equal("Currently guessing AC-2");
  });

  it("should return correct roundStateAsReadableString when round is done", () => {
    // @ts-ignore
    estimationStore.roomState = defaultRoomState
      .withRoundState(RoundState.End)
      .withIssueToGuess("AC-2")
      .build();
    const wrapper = createWrapper();

    // @ts-ignore
    expect(wrapper.vm.roundStateAsReadableString).equal("Everyone guessed!");
  });
});

function createWrapper() {
  return vuetifyMount(RoomView, {
    global: {
      plugins: [pinia],
      stubs: {
        VNavigationDrawer: {
          template: "<div></div>",
        },
      },
    },
  });
}
