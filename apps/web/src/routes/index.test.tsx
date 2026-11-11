import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { createInMemoryHealthGateway } from '#/features/health/in-memory-health-gateway'
import { renderRoute } from '#/test/renderRoute'

describe('Homepage', () => {
  it('says the API is reachable when the health check answers', async () => {
    await renderRoute('/')

    expect(
      screen.getByRole('heading', { name: /buy and sell second-hand items/i }),
    ).toBeInTheDocument()
    expect(await screen.findByText('API reachable')).toBeInTheDocument()
  })

  it('says the API is unreachable when the health check fails', async () => {
    await renderRoute('/', {
      gateways: { health: createInMemoryHealthGateway({ reachable: false }) },
    })

    expect(await screen.findByText('API unreachable')).toBeInTheDocument()
  })
})
