import { useAppForm } from '#/hooks/form'
import { motifSchema } from '#/features/moderation/schemas'

// Asks the moderator why the vendeur must correct the annonce: the motif is
// emailed to the vendeur, with the link to the correction.
export function ChallengerForm({
  onSubmit,
}: {
  onSubmit: (motif: string) => void
}) {
  const form = useAppForm({
    defaultValues: { motif: '' },
    validators: { onChange: motifSchema },
    onSubmit: ({ value }) => onSubmit(value.motif.trim()),
  })

  return (
    <form
      noValidate
      className="flex max-w-sz-480 flex-col gap-md"
      onSubmit={(event) => {
        event.preventDefault()
        event.stopPropagation()
        void form.handleSubmit()
      }}
    >
      <form.AppField name="motif">
        {(field) => (
          <field.TextArea
            label="Motif du challenge"
            description="De 10 à 1 000 caractères, envoyé au vendeur avec le lien de correction."
            rows={4}
            isRequired
          />
        )}
      </form.AppField>
      <form.AppForm>
        <form.SubscribeButton label="Challenger" />
      </form.AppForm>
    </form>
  )
}
