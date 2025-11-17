import { defineStore } from "pinia";

export const useDarkModeStore = defineStore("darkMode", {
  state: () => ({
    isDark: localStorage.getItem("isDarkMode") === "1",
  }),
  actions: {
    setDark(val: boolean) {
      this.isDark = val;
      localStorage.setItem("isDarkMode", val ? "1" : "0");
    },
    toggle() {
      this.setDark(!this.isDark);
    },
  },
});
