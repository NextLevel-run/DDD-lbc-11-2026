import { useAppForm } from '#/hooks/form'
import { moderateurSchema } from '#/features/moderation/schemas'

// Asks the moderator for his pseudo: there is no account, the pseudo is what
// the locks and the history carry.
export function PseudoModerateurForm({
  onSubmit,
}: {
  onSubmit: (pseudo: string) => void
}) {
  const form = useAppForm({
    defaultValues: { moderateur: '' },
    validators: { onChange: moderateurSchema },
    onSubmit: ({ value }) => onSubmit(value.moderateur.trim()),
  })

  return (
    <form
      noValidate
      className="flex max-w-sz-480 flex-col gap-lg"
      onSubmit={(event) => {
        event.preventDefault()
        event.stopPropagation()
        void form.handleSubmit()
      }}
    >
      <form.AppField name="moderateur">
        {(field) => (
          <field.TextField
            label="Votre pseudo de modérateur"
            description="De 2 à 30 caractères, conservé dans ce navigateur."
            autoComplete="nickname"
            isRequired
          />
        )}
      </form.AppField>
      <form.AppForm>
        <form.SubscribeButton label="Continuer" />
      </form.AppForm>
    </form>
  )
}
