import { describe, expect, it } from "vitest";
import router from "../../src/router";

describe("router", () => {
  it("should be called correctly", () => {
    expect(router.getRoutes()[0].path).equal("/");
    expect(router.getRoutes()[0].name).equal("home");
    // @ts-ignore
    expect(router.getRoutes()[0].components.default.__name).equal("HomeView");

    expect(router.getRoutes()[1].path).equal("/room");
    expect(router.getRoutes()[1].name).equal("room");
  });
});
