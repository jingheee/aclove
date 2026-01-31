import { createRouter, createWebHistory } from "vue-router";
import Home from "@/components/Home.vue";
import PostList from "@/components/post/PostList.vue";
import PostDetail from "@/components/post/PostDetail.vue";

const routes = [
  {
    path: "/",
    name: "home",
    component: Home,
  },
  {
    path: "/category/:categoryId",
    name: "category",
    component: PostList,
    props: true,
  },
  {
    path: "/post/:postId",
    name: "post-detail",
    component: PostDetail,
    props: true,
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
