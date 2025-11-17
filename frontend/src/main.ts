import { createApp } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import "mdui/mdui.css";

import App from "./App.vue";
import ControlPanel from "./views/ControlPanel.vue";
import ShowSource from "./views/ShowSource.vue";
import KeyManagement from "./views/KeyManagement.vue";
import { createPinia } from "pinia";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      redirect: "/control-panel",
    },
    {
      path: "/control-panel",
      name: "control-panel",
      component: ControlPanel,
    },
    {
      path: "/control-panel/keys",
      name: "key-management",
      component: KeyManagement,
    },
    {
      path: "/show-source",
      name: "show-source",
      component: ShowSource,
    },
  ],
});

const pinia = createPinia();

createApp(App).use(router).use(pinia).mount("#app");
