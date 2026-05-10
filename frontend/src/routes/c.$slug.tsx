import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/c/$slug')({
  component: RouteComponent,
})

function RouteComponent() {
  const { slug } = Route.useParams()

  return (
    <div className="flex items-center justify-center min-h-screen">
      <p className="text-body text-neutral-600">Chat público: {slug}</p>
    </div>
  )
}
