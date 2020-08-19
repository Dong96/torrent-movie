import Vue from "vue";
import App from "./App";
import router from "./router";
import CoreuiVue from "@coreui/vue";
import store from "./store/store";
import axios from "axios";

Vue.config.performance = true;
Vue.use(CoreuiVue);
Vue.prototype.$log = console.log.bind(console);

axios.defaults.baseURL = "http://localhost:8081";
axios.defaults.timeout = 2500;

new Vue({
  el: "#app",
  router,
  store,
  template: "<App/>",
  components: {
    App,
  },
  render: (h) => h(App),
});
