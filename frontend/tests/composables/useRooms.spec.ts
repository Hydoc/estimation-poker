import { describe, it, expect, vi } from "vitest";
import { useRooms } from "../../src/composables/useRooms";
import { fail, none, succeed } from "@kaumlaut/pure/fetch-state";

describe("useRooms", () => {
  describe("state", () => {
    it("should have default state", () => {
      const composable = useRooms();

      expect(composable.roomsState.value).deep.equal({
        availableActiveRooms: none(),
      });
    });
  });

  describe("fetchActiveRooms", () => {
    it("should fetch and set", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: true,
        json: () =>
          Promise.resolve({
            rooms: [
              { id: "any-id", playerCount: 1 },
              { id: "any-id-2", playerCount: 5 },
            ],
          }),
      }));
      const composable = useRooms();

      await composable.fetchActiveRooms();

      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/rooms");
      expect(composable.roomsState.value.availableActiveRooms).deep.equal(
        succeed({
          rooms: [
            { id: "any-id", playerCount: 1 },
            { id: "any-id-2", playerCount: 5 },
          ],
        }),
      );
    });

    it("should fail when response not ok", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: false,
      }));
      const composable = useRooms();

      await composable.fetchActiveRooms();

      expect(composable.roomsState.value.availableActiveRooms).deep.equal(
        fail("error fetching active rooms"),
      );
    });

    it("should fail when incorrect type was returned", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: true,
        json: () =>
          Promise.resolve({
            invalid: true,
          }),
      }));
      const composable = useRooms();

      await composable.fetchActiveRooms();

      expect(composable.roomsState.value.availableActiveRooms).deep.equal(
        fail("Object does not have key rooms"),
      );
    });
  });

  describe("createRoom", () => {
    it("should create", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: true,
        json: () =>
          Promise.resolve({
            id: "created-id",
          }),
      }));
      const composable = useRooms();

      const result = await composable.createRoom("Tester");

      expect(global.fetch).toHaveBeenNthCalledWith(1, "/v1/room", {
        method: "POST",
        body: JSON.stringify({
          creator: "Tester",
          guesses: {},
        }),
      });
      expect(result).equal("created-id");
    });

    it("should throw error when creating fails", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: false,
      }));
      const composable = useRooms();

      await expect(() => composable.createRoom("Tester")).rejects.toThrow("could not create room");
    });

    it("should throw error when response is of incorrect type", async () => {
      // @ts-ignore
      global.fetch = vi.fn(() => ({
        ok: true,
        json: () =>
          Promise.resolve({
            type: "incorrect",
          }),
      }));
      const composable = useRooms();

      await expect(() => composable.createRoom("Tester")).rejects.toThrow("could not create room");
    });
  });
});
