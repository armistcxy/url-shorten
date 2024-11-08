import Home from "./pages/Home";
import Preview from "./pages/Preview";
import NotFound from "./pages/NotFound";

const routes = [
  {
    path: "",
    component: Home,
  },
  {
    path: "/preview",
    component: Preview,
  },
  {
    path: "*",
    component: NotFound,
  },
];

export default routes;
