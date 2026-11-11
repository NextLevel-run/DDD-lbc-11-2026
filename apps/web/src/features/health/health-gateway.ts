export interface Health {
  status: 'ok'
}

// Port: what the UI needs from the API to know it is reachable.
export interface HealthGateway {
  check: () => Promise<Health>
}
