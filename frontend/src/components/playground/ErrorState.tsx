import { Banner } from "@cloudflare/kumo/components/banner"
import type { ApiError } from "@/types/api"

type ErrorStateProps = {
  error: Pick<ApiError, "kind" | "message">
  onRetry: () => void
}

const messages: Record<ApiError["kind"], { variant: "default" | "alert" | "error" | "secondary"; title: string; text: string }> = {
  network: {
    variant: "error",
    title: "Network error",
    text: "Could not reach the API. Check your connection and try again.",
  },
  timeout: {
    variant: "alert",
    title: "Request timed out",
    text: "The request took too long to complete. Please try again.",
  },
  rate_limited: {
    variant: "alert",
    title: "Rate limit reached",
    text: "You have reached the hourly request limit (10 requests/hour). Please try again later.",
  },
  unauthorized: {
    variant: "error",
    title: "Invalid API key",
    text: "The API key configured for this playground is invalid.",
  },
  server: {
    variant: "error",
    title: "API unavailable",
    text: "The API is temporarily unavailable. Please try again.",
  },
  client: {
    variant: "error",
    title: "Request failed",
    text: "The request could not be processed.",
  },
  unknown: {
    variant: "error",
    title: "Unexpected error",
    text: "An unexpected error occurred.",
  },
}

export default function ErrorState({ error, onRetry }: ErrorStateProps) {
  const config = messages[error.kind]
  const canRetry = error.kind !== "rate_limited"
  const description =
    error.kind === "client" && error.message ? `${config.text} ${error.message}` : config.text

  return (
    <Banner
      variant={config.variant}
      title={config.title}
      description={description}
      action={
        canRetry ? (
          <Banner.Action variant="secondary" onClick={onRetry}>
            Try again
          </Banner.Action>
        ) : undefined
      }
    />
  )
}
