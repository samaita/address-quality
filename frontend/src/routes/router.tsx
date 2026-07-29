import { createBrowserRouter } from "react-router-dom"
import MainLayout from "@/layouts/MainLayout"
import Landing from "@/pages/Landing"
import Playground from "@/pages/Playground"
import Docs from "@/pages/Docs"
import NotFound from "@/pages/NotFound"

const router = createBrowserRouter([
  {
    element: <MainLayout />,
    children: [
      { index: true, element: <Landing /> },
      { path: "playground", element: <Playground /> },
      { path: "docs", element: <Docs /> },
      { path: "*", element: <NotFound /> },
    ],
  },
])

export default router
