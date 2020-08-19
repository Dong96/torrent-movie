import Vue from "vue";
import Router from "vue-router";

// Containers
const TheContainer = () => import("@/container/TheContainer");

// Views
const Table = () => import("@/view/datatable/Table");

const Player = () => import("@/view/player/Player");

Vue.use(Router);

export default new Router({
  mode: "history", // https://router.vuejs.org/api/#mode
  linkActiveClass: "active",
  scrollBehavior: () => ({ y: 0 }),
  routes: configRoutes(),
});

function configRoutes() {
  return [
    {
      path: "/",
      redirect: "/",
      name: "Home",
      component: TheContainer,
      children: [
        {
          path: "/",
          name: "Table",
          component: Table,
        },
        {
          path: "/movie/:movie",
          name: "Player",
          component: Player,
          props: true,
        },
      ],
    },
  ];
}
