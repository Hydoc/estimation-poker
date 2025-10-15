import { describe, expect, it } from "vitest";
import { VIcon } from "vuetify/components";
import DeveloperCard from "../../src/components/DeveloperCard.vue";
import { Developer, DeveloperDone, Role } from "../../src/components/types";
import { vuetifyMount } from "../vuetifyMount";

describe("DeveloperCard", () => {
  describe("rendering", () => {
    it("should render when developer not done", () => {
      const wrapper = createWrapper();
      expect(wrapper.find(".developer-card > span").text()).equal("Test Dev");
      expect(wrapper.find(".reveal").exists()).to.be.false;
      expect(wrapper.find(".waiting-for-guess").exists()).to.be.true;
      expect(wrapper.find(".guessed").exists()).to.be.false;
      expect(wrapper.find(".flip-card__back > span").exists()).to.be.false;
    });

    it("should render when developer is done but it has not been revealed", () => {
      const wrapper = createWrapper({
        name: "Test Dev",
        isDone: true,
        role: Role.Developer,
      });
      expect(wrapper.find(".reveal").exists()).to.be.false;
      expect(wrapper.find(".waiting-for-guess").exists()).to.be.false;
      expect(wrapper.find(".guessed").exists()).to.be.true;
      expect(wrapper.find(".flip-card__back > span").exists()).to.be.false;
    });

    it("should render guess when round has been revealed", () => {
      const dev: Developer = {
        name: "Test Dev",
        isDone: true,
        role: Role.Developer,
      };
      const wrapper = createWrapper(dev, {
        ...dev,
        doSkip: false,
        guess: 2,
      });
      expect(wrapper.find(".reveal").exists()).to.be.true;
      expect(wrapper.find(".waiting-for-guess").exists()).to.be.false;
      expect(wrapper.find(".guessed").exists()).to.be.true;
      expect(wrapper.find(".flip-card__back > span").exists()).to.be.true;
      expect(wrapper.find(".flip-card__back > span").text()).equal("2");
    });

    it("should render skip icon when round has been revealed and developer skipped", () => {
      const dev: Developer = {
        name: "Test Dev",
        isDone: true,
        role: Role.Developer,
      };
      const wrapper = createWrapper(dev, {
        ...dev,
        doSkip: true,
        guess: 0,
      });
      expect(wrapper.find(".reveal").exists()).to.be.true;
      expect(wrapper.find(".waiting-for-guess").exists()).to.be.false;
      expect(wrapper.find(".guessed").exists()).to.be.true;
      expect(wrapper.findComponent(VIcon).exists()).to.be.true;
      expect(wrapper.findComponent(VIcon).find("i").classes()).contains("mdi-coffee");
    });
  });
});

function createWrapper(
  developer: Developer = { name: "Test Dev", isDone: false, role: Role.Developer },
  developerDone: DeveloperDone | undefined = undefined,
) {
  return vuetifyMount(DeveloperCard, {
    props: {
      developer,
      developerDone,
    },
  });
}
