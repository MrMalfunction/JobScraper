import { createRouter, createWebHistory } from "vue-router";
import Dashboard from "../views/Dashboard.vue";
import TodaysJobs from "../views/TodaysJobs.vue";
import AllJobs from "../views/AllJobs.vue";
import Companies from "../views/Companies.vue";

const routes = [
  {
    path: "/",
    name: "Dashboard",
    component: Dashboard,
  },
  {
    path: "/todays-jobs",
    name: "TodaysJobs",
    component: TodaysJobs,
  },
  {
    path: "/all-jobs",
    name: "AllJobs",
    component: AllJobs,
  },
  {
    path: "/companies",
    name: "Companies",
    component: Companies,
  },
];

const router = createRouter({
  history: createWebHistory("/ui/"),
  routes,
});

export default router;
