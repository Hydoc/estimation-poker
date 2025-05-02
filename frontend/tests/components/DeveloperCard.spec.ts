import { beforeEach, describe, expect, it } from "vitest";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { mount } from "@vue/test-utils";
import DeveloperCard from "../../src/components/DeveloperCard.vue";
import { Developer, DeveloperDone } from "../../src/components/types";
import { VIcon } from "vuetify/components";

let vuetify: ReturnType<typeof createVuetify>;

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});
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
        role: "developer",
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
        role: "developer",
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
        role: "developer",
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
  developer: Developer = { name: "Test Dev", isDone: false, role: "developer" },
  developerDone: DeveloperDone | undefined = undefined,
) {
  return mount(DeveloperCard, {
    props: {
      developer,
      developerDone,
    },
    global: {
      plugins: [vuetify],
    },
  });
}
