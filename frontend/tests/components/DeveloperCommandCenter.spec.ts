import { describe, expect, it } from "vitest";
import DeveloperCommandCenter from "../../src/components/DeveloperCommandCenter.vue";
import { VIcon } from "vuetify/components";
import { vuetifyMount } from "../vuetifyMount";

describe("DeveloperCommandCenter", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = vuetifyMount(DeveloperCommandCenter, {
        props: {
          guess: 0,
          showAllGuesses: false,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Up to 4h" },
            { guess: 2, description: "Up to 8h" },
            { guess: 3, description: "Up to 3 days" },
            { guess: 4, description: "Up to 5 days" },
            { guess: 5, description: "More than 5 days" },
          ],
        },
      });

      expect(wrapper.findAll(".card")).length(6);
      expect(wrapper.findAll(".card").at(0).find("h2").text()).equal("1");
      expect(wrapper.findAll(".card").at(0).find("span").text()).equal("Up to 4h");

      expect(wrapper.findAll(".card").at(1).find("h2").text()).equal("2");
      expect(wrapper.findAll(".card").at(1).find("span").text()).equal("Up to 8h");

      expect(wrapper.findAll(".card").at(2).find("h2").text()).equal("3");
      expect(wrapper.findAll(".card").at(2).find("span").text()).equal("Up to 3 days");

      expect(wrapper.findAll(".card").at(3).find("h2").text()).equal("4");
      expect(wrapper.findAll(".card").at(3).find("span").text()).equal("Up to 5 days");

      expect(wrapper.findAll(".card").at(4).find("h2").text()).equal("5");
      expect(wrapper.findAll(".card").at(4).find("span").text()).equal("More than 5 days");

      expect(wrapper.findAll(".card").at(5).findComponent(VIcon).find("i").classes()).contains(
        "mdi-coffee",
      );
      expect(wrapper.findAll(".card").at(5).find("span").text()).equal("Skip round");
    });

    it("should render with correct guess when developer did guess", async () => {
      const wrapper = vuetifyMount(DeveloperCommandCenter, {
        props: {
          guess: 2,
          showAllGuesses: false,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Up to 4h" },
            { guess: 2, description: "Up to 8h" },
            { guess: 3, description: "Up to 3 days" },
            { guess: 4, description: "Up to 5 days" },
            { guess: 5, description: "More than 5 days" },
          ],
        },
      });

      expect(wrapper.find(".active-guess").find("h2").text()).equal("2");
      expect(wrapper.find(".active-guess").find("span").text()).equal("Up to 8h");
    });

    it("should render with correct color when developer didSkip", () => {
      const wrapper = vuetifyMount(DeveloperCommandCenter, {
        props: {
          guess: 0,
          showAllGuesses: false,
          didSkip: true,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Up to 4h" },
            { guess: 2, description: "Up to 8h" },
            { guess: 3, description: "Up to 3 days" },
            { guess: 4, description: "Up to 5 days" },
            { guess: 5, description: "More than 5 days" },
          ],
        },
      });

      expect(wrapper.find(".active-guess").findComponent(VIcon).find("i").classes()).contains(
        "mdi-coffee",
      );
      expect(wrapper.find(".active-guess").find("span").text()).equal("Skip round");
    });
  });

  describe("functionality", () => {
    it("should emit guess on card click submit", async () => {
      const wrapper = vuetifyMount(DeveloperCommandCenter, {
        props: {
          guess: 0,
          showAllGuesses: false,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Up to 4h" },
            { guess: 2, description: "Up to 8h" },
            { guess: 3, description: "Up to 3 days" },
            { guess: 4, description: "Up to 5 days" },
            { guess: 5, description: "More than 5 days" },
          ],
        },
      });

      await wrapper.findAll(".card").at(2).trigger("click");
      expect(wrapper.emitted("guess")).deep.equal([[3]]);

      await wrapper.findAll(".card").at(3).trigger("click");
      expect(wrapper.emitted("guess")).deep.equal([[3], [4]]);
    });

    it("should emit skip on skip card press", async () => {
      const wrapper = vuetifyMount(DeveloperCommandCenter, {
        props: {
          guess: 0,
          showAllGuesses: false,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Up to 4h" },
            { guess: 2, description: "Up to 8h" },
            { guess: 3, description: "Up to 3 days" },
            { guess: 4, description: "Up to 5 days" },
            { guess: 5, description: "More than 5 days" },
          ],
        },
      });

      await wrapper.findAll(".card").at(5).trigger("click");
      expect(wrapper.emitted("skip")).deep.equal([[]]);
    });
  });
});
