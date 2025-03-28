import { beforeEach, describe, expect, it } from "vitest";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import { VIcon, VTable } from "vuetify/components";
import * as directives from "vuetify/directives";
import { mount } from "@vue/test-utils";
import ResultTable from "../../src/components/ResultTable.vue";
import { Role } from "../../src/components/types";

let vuetify: ReturnType<typeof createVuetify>;

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});
describe("ResultTable", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = mount(ResultTable, {
        props: {
          developerDone: [],
          developerList: [
            { name: "test", isDone: false, role: Role.Developer },
            { name: "another", isDone: true, role: Role.Developer },
          ],
          showAllGuesses: false,
          roundIsFinished: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VTable).exists()).to.be.true;
      expect(wrapper.findComponent(VTable).findAll("thead th")).length(2);
      expect(wrapper.findComponent(VTable).findAll("thead th").at(0).text()).equal("Name");
      expect(wrapper.findComponent(VTable).findAll("thead th").at(1).text()).equal("Schätzung");
      // because two developers in room
      expect(wrapper.findAll("tbody tr")).length(2);

      // first developer
      expect(wrapper.findAll("tbody tr").at(0).findAll("td").at(0).text()).equal("test");
      expect(
        wrapper
          .findAll("tbody tr")
          .at(0)
          .findAll("td")
          .at(1)
          .findComponent(VIcon)
          .find("i")
          .classes(),
      ).contains("mdi-help-circle");

      // second developer
      expect(wrapper.findAll("tbody tr").at(1).findAll("td").at(0).text()).equal("another");
      expect(
        wrapper
          .findAll("tbody tr")
          .at(1)
          .findAll("td")
          .at(1)
          .findComponent(VIcon)
          .find("i")
          .classes(),
      ).contains("mdi-check-circle");
      expect(
        wrapper.findAll("tbody tr").at(1).findAll("td").at(1).findComponent(VIcon).props("color"),
      ).equal("green");
    });

    it("should return 0 as average guess when there a no devs that have guessed", () => {
      const wrapper = mount(ResultTable, {
        props: {
          developerDone: [{ name: "test", guess: 0, role: Role.Developer, doSkip: true }],
          developerList: [{ name: "test", isDone: true, role: Role.Developer }],
          showAllGuesses: true,
          roundIsFinished: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findAll("tbody tr").at(0).findAll("td").at(0).text()).equal("test");
      expect(wrapper.findAll("tbody tr").at(0).findAll("td").at(1).find(".mdi-coffee").exists())
        .true;
      expect(wrapper.findAll("tbody tr").at(1).findAll("td").at(0).text()).equal("Durchschnitt");
      expect(wrapper.findAll("tbody tr").at(1).findAll("td").at(1).text()).equal("0");
    });

    it("should render all guesses and average as text when show all guesses = true", () => {
      const wrapper = mount(ResultTable, {
        props: {
          developerDone: [
            { name: "test", guess: 2, doSkip: false, role: Role.Developer },
            { name: "another 1", guess: 1, doSkip: false, role: Role.Developer },
            { name: "another 2", guess: 1, doSkip: false, role: Role.Developer },
            { name: "another 3", guess: 1, doSkip: false, role: Role.Developer },
            { name: "another 4", guess: 2, doSkip: false, role: Role.Developer },
          ],
          developerList: [
            { name: "test", isDone: true, role: Role.Developer },
            { name: "another 1", isDone: true, role: Role.Developer },
            { name: "another 2", isDone: true, role: Role.Developer },
            { name: "another 3", isDone: true, role: Role.Developer },
            { name: "another 4", isDone: true, role: Role.Developer },
          ],
          showAllGuesses: true,
          roundIsFinished: false,
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findAll("tbody tr").at(0).findAll("td").at(1).text()).equal("2");
      expect(wrapper.findAll("tbody tr").at(0).classes()).contains("bg-blue-grey-lighten-5");
      expect(wrapper.findAll("tbody tr").at(1).findAll("td").at(1).text()).equal("1");
      expect(wrapper.findAll("tbody tr").at(2).findAll("td").at(1).text()).equal("1");
      expect(wrapper.findAll("tbody tr").at(3).findAll("td").at(1).text()).equal("1");
      expect(wrapper.findAll("tbody tr").at(4).findAll("td").at(1).text()).equal("2");
      expect(wrapper.findAll("tbody tr").at(4).classes()).contains("bg-blue-grey-lighten-5");
      expect(wrapper.findAll("tbody tr").at(5).findAll("td").at(0).text()).equal("Durchschnitt");
      expect(wrapper.findAll("tbody tr").at(5).findAll("td").at(1).text()).equal("1");
    });
  });
});
