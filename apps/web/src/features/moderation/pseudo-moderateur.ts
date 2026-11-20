import { useEffect, useState } from 'react'

const CLE = 'moderateur'

// The pseudo the moderator typed, kept in localStorage. `undefined` until it
// has been read (the server renders without it), empty when there is none.
export function usePseudoModerateur(): [
  string | undefined,
  (pseudo: string) => void,
] {
  const [pseudo, setPseudo] = useState<string | undefined>(undefined)

  useEffect(() => {
    setPseudo(window.localStorage.getItem(CLE) ?? '')
  }, [])

  return [
    pseudo,
    (nouveau) => {
      window.localStorage.setItem(CLE, nouveau)
      setPseudo(nouveau)
    },
  ]
}
