import { describe, it, expect, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import UserBox from "../../src/components/UserBox.vue";
import { User } from "../../src/components/types";
import { createVuetify } from "vuetify";
import { VCard, VCardTitle, VList, VListItem } from "vuetify/components";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";

let vuetify: ReturnType<typeof createVuetify>;

const userList: User[] = [
  {
    name: "One",
  },
  {
    name: "Two",
  },
];

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});
describe("UserBox", () => {
  describe("rendering", () => {
    it("should render", () => {
      const wrapper = mount(UserBox, {
        props: {
          userList,
          currentUsername: "One",
          title: "Developer",
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VCard).exists()).to.be.true;
      expect(wrapper.findComponent(VCardTitle).text()).equal("Developer");
      expect(wrapper.findComponent(VList).exists()).to.be.true;
      expect(wrapper.findAllComponents(VListItem)).length(2);
      expect(wrapper.findAllComponents(VListItem).at(0).text()).equal("One (Du)");
      expect(wrapper.findAllComponents(VListItem).at(1).text()).equal("Two");
    });

    it("should render with empty userList", () => {
      const wrapper = mount(UserBox, {
        props: {
          userList: [],
          currentUsername: "One",
          title: "Developer",
        },
        global: {
          plugins: [vuetify],
        },
      });

      expect(wrapper.findComponent(VList).exists()).to.be.false;
    });
  });
});
