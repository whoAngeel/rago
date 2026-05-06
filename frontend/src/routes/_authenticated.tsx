import { createFileRoute, redirect } from '@tanstack/react-router'
import { AppLayout } from '../components/layout/AppLayout'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async () => {
    const { accessToken } = await import('../store/authStore').then(m => m.useAuthStore.getState())
    if (!accessToken) {
      throw redirect({
        to: '/login',
        search: {
          redirect: window.location.pathname
        }
      })
    }

  },
  component: AppLayout,
})

