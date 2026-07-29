import Card from "@/components/common/Card"
import Badge from "@/components/common/Badge"
import CodeBlock from "@/components/docs/CodeBlock"

type EndpointCardProps = {
  method: string
  path: string
  description: string
  requestCode: string
  responseCode: string
  errorCode?: string
}

export default function EndpointCard({
  method,
  path,
  description,
  requestCode,
  responseCode,
  errorCode,
}: EndpointCardProps) {
  return (
    <Card className="space-y-6">
      <div className="flex items-center gap-3">
        <Badge variant="info">{method}</Badge>
        <code className="font-mono text-sm text-surface-900">{path}</code>
      </div>
      <p className="text-sm leading-relaxed text-surface-600">{description}</p>
      <div className="space-y-4">
        <div>
          <p className="mb-2 text-xs font-medium uppercase tracking-wider text-surface-500">
            Request
          </p>
          <CodeBlock code={requestCode} title={`${method} ${path}`} />
        </div>
        <div>
          <p className="mb-2 text-xs font-medium uppercase tracking-wider text-surface-500">
            Response
          </p>
          <CodeBlock code={responseCode} title="200 OK" />
        </div>
        {errorCode && (
          <div>
            <p className="mb-2 text-xs font-medium uppercase tracking-wider text-surface-500">
              Error
            </p>
            <CodeBlock code={errorCode} title="400 Bad Request" />
          </div>
        )}
      </div>
    </Card>
  )
}
