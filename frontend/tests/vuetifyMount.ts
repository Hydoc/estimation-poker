import { beforeEach } from "vitest";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { createVuetify } from "vuetify";
import { mount } from "@vue/test-utils";

let vuetify: ReturnType<typeof createVuetify>;

beforeEach(() => {
  vuetify = createVuetify({
    components,
    directives,
  });
});

export function vuetifyMount<T>(component: T, opts: Record<string, any>) {
  return mount(component, {
    ...opts,
    global: {
      ...(opts.global || {}),
      plugins: [...(opts.global || { plugins: [] }).plugins, vuetify],
    },
  });
}
