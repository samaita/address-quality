import Container from "@/components/layout/Container"
import Button from "@/components/common/Button"

export default function NotFound() {
  return (
    <Container className="flex flex-col items-center justify-center py-24 text-center">
      <p className="text-6xl font-semibold text-surface-300">404</p>
      <h1 className="mt-4 text-2xl font-semibold text-surface-900">
        Page not found
      </h1>
      <p className="mt-2 text-surface-500">
        The page you are looking for does not exist.
      </p>
      <Button as="a" href="/" variant="primary" className="mt-8">
        Back to home
      </Button>
    </Container>
  )
}
