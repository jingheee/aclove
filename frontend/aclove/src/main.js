import NaiveUI, { createDiscreteApi } from "naive-ui";
import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import { VueQueryPlugin } from "./queryClient";
import "./style.css";

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(NaiveUI);
app.use(VueQueryPlugin);

const { message, dialog } = createDiscreteApi(["message", "dialog"]);

app.config.globalProperties.$message = message;
app.config.globalProperties.$dialog = dialog;

app.mount("#app");
