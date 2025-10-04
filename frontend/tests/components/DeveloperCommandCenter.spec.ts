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
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      expect(wrapper.findAll(".card")).length(6);
      expect(wrapper.findAll(".card").at(0).find("h2").text()).equal("1");
      expect(wrapper.findAll(".card").at(0).find("span").text()).equal("Bis zu 4 Std.");

      expect(wrapper.findAll(".card").at(1).find("h2").text()).equal("2");
      expect(wrapper.findAll(".card").at(1).find("span").text()).equal("Bis zu 8 Std.");

      expect(wrapper.findAll(".card").at(2).find("h2").text()).equal("3");
      expect(wrapper.findAll(".card").at(2).find("span").text()).equal("Bis zu 3 Tagen");

      expect(wrapper.findAll(".card").at(3).find("h2").text()).equal("4");
      expect(wrapper.findAll(".card").at(3).find("span").text()).equal("Bis zu 5 Tagen");

      expect(wrapper.findAll(".card").at(4).find("h2").text()).equal("5");
      expect(wrapper.findAll(".card").at(4).find("span").text()).equal("Mehr als 5 Tage");

      expect(wrapper.findAll(".card").at(5).findComponent(VIcon).find("i").classes()).contains(
        "mdi-coffee",
      );
      expect(wrapper.findAll(".card").at(5).find("span").text()).equal("Runde aussetzen");
    });

    it("should render with correct guess when developer did guess", async () => {
      const wrapper = vuetifyMount(DeveloperCommandCenter, {
        props: {
          guess: 2,
          showAllGuesses: false,
          didSkip: false,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      expect(wrapper.find(".active-guess").find("h2").text()).equal("2");
      expect(wrapper.find(".active-guess").find("span").text()).equal("Bis zu 8 Std.");
    });

    it("should render with correct color when developer didSkip", () => {
      const wrapper = vuetifyMount(DeveloperCommandCenter, {
        props: {
          guess: 0,
          showAllGuesses: false,
          didSkip: true,
          hasTicketToGuess: true,
          possibleGuesses: [
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      expect(wrapper.find(".active-guess").findComponent(VIcon).find("i").classes()).contains(
        "mdi-coffee",
      );
      expect(wrapper.find(".active-guess").find("span").text()).equal("Runde aussetzen");
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
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
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
            { guess: 1, description: "Bis zu 4 Std." },
            { guess: 2, description: "Bis zu 8 Std." },
            { guess: 3, description: "Bis zu 3 Tagen" },
            { guess: 4, description: "Bis zu 5 Tagen" },
            { guess: 5, description: "Mehr als 5 Tage" },
          ],
        },
      });

      await wrapper.findAll(".card").at(5).trigger("click");
      expect(wrapper.emitted("skip")).deep.equal([[]]);
    });
  });
});
