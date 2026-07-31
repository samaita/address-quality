import { forwardRef } from "react"
import { Link as ReactRouterLink, RouterProvider } from "react-router-dom"
import { LinkProvider, type LinkComponentProps } from "@cloudflare/kumo/utils"
import router from "@/routes/router"

const AppLink = forwardRef<HTMLAnchorElement, LinkComponentProps>(
  ({ href, ...rest }, ref) => <ReactRouterLink ref={ref} to={href ?? ""} {...rest} />,
)

export default function App() {
  return (
    <LinkProvider component={AppLink}>
      <RouterProvider router={router} />
    </LinkProvider>
  )
}
