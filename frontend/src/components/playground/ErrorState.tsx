import { Button } from "@cloudflare/kumo/components/button"
import { Alert } from "@/components/kumo-ui"
import type { ApiError } from "@/types/api"

type ErrorStateProps = {
  error: Pick<ApiError, "kind" | "message">
  onRetry: () => void
}

const messages: Record<ApiError["kind"], { variant: "info" | "success" | "warning" | "danger"; title: string; text: string }> = {
  network: {
    variant: "danger",
    title: "Network error",
    text: "Could not reach the API. Check your connection and try again.",
  },
  timeout: {
    variant: "warning",
    title: "Request timed out",
    text: "The request took too long to complete. Please try again.",
  },
  rate_limited: {
    variant: "warning",
    title: "Rate limit reached",
    text: "You have reached the hourly request limit (10 requests/hour). Please try again later.",
  },
  unauthorized: {
    variant: "danger",
    title: "Invalid API key",
    text: "The API key configured for this playground is invalid.",
  },
  server: {
    variant: "danger",
    title: "API unavailable",
    text: "The API is temporarily unavailable. Please try again.",
  },
  client: {
    variant: "danger",
    title: "Request failed",
    text: "The request could not be processed.",
  },
  unknown: {
    variant: "danger",
    title: "Unexpected error",
    text: "An unexpected error occurred.",
  },
}

export default function ErrorState({ error, onRetry }: ErrorStateProps) {
  const config = messages[error.kind]
  const canRetry = error.kind !== "rate_limited"

  return (
    <Alert
      variant={config.variant}
      title={config.title}
      action={
        canRetry ? (
          <Button type="button" variant="secondary" size="sm" onClick={onRetry}>
            Try again
          </Button>
        ) : undefined
      }
    >
      {config.text}
      {error.kind === "client" && error.message ? ` ${error.message}` : ""}
    </Alert>
  )
}
