import { describe, expect, it } from 'vitest'
import {
  centimesEnEuros,
  corrigerAnnonceSchema,
  deposerAnnonceSchema,
  eurosEnCentimes,
  criteresDeRecherche,
  rechercheSearchSchema,
} from './schemas'
import type { DeposerAnnonceValues } from './schemas'

const valid: DeposerAnnonceValues = {
  titre: 'Vélo de course',
  description: 'Cadre carbone, très bon état.',
  prix: '450',
  categorie: 'loisirs',
  ville: 'Lyon',
  codePostal: '69003',
  pseudo: 'cycliste69',
  email: 'cycliste@example.com',
  telephone: '',
  motDePasse: 'demo1234',
}

const messagesFor = (values: Partial<DeposerAnnonceValues>) => {
  const result = deposerAnnonceSchema.safeParse({ ...valid, ...values })
  return result.success ? [] : result.error.issues.map((issue) => issue.message)
}

describe('centimesEnEuros', () => {
  it.each([
    [0, '0'],
    [45000, '450'],
    [45050, '450,50'],
    [5, '0,05'],
  ])('writes %d centimes as %s', (centimes, euros) => {
    expect(centimesEnEuros(centimes)).toBe(euros)
    expect(eurosEnCentimes(euros)).toBe(centimes)
  })
})

describe('eurosEnCentimes', () => {
  it.each([
    ['0', 0],
    ['450', 45000],
    ['12,5', 1250],
    ['12.50', 1250],
    ['0,99', 99],
    [' 1 250 ', 125000],
  ])('reads %s euros as %i centimes', (euros, centimes) => {
    expect(eurosEnCentimes(euros)).toBe(centimes)
  })

  it.each(['', 'abc', '-3', '1,234', '1e3'])('rejects %s', (euros) => {
    expect(eurosEnCentimes(euros)).toBeUndefined()
  })
})

describe('deposerAnnonceSchema mot de passe', () => {
  it('refuses fewer than 8 characters', () => {
    expect(messagesFor({ motDePasse: 'court' })).toEqual([
      "Le mot de passe de l'annonce doit comporter au moins 8 caractères",
    ])
    expect(messagesFor({ motDePasse: '' })).toHaveLength(1)
  })

  it('accepts 8 characters with no other rule and keeps it untrimmed', () => {
    expect(
      deposerAnnonceSchema.parse({ ...valid, motDePasse: 'aaaaaaaa' })
        .motDePasse,
    ).toBe('aaaaaaaa')
  })
})

describe('deposerAnnonceSchema', () => {
  it('turns valid values into the args of the dépôt', () => {
    expect(
      deposerAnnonceSchema.parse({
        ...valid,
        titre: '  Vélo de course ',
        prix: '12,50',
        telephone: '06 12 34 56 78',
      }),
    ).toEqual({
      ...valid,
      prix: 1250,
      telephone: '0612345678',
    })
  })

  it('accepts a don', () => {
    expect(deposerAnnonceSchema.parse({ ...valid, prix: '0' }).prix).toBe(0)
  })

  it.each([
    [{ titre: 'ab' }, 'Le titre doit comporter de 3 à 100 caractères'],
    [
      { titre: 'a'.repeat(101) },
      'Le titre doit comporter de 3 à 100 caractères',
    ],
    [
      { description: 'court' },
      'La description doit comporter de 10 à 4 000 caractères',
    ],
    [{ prix: '' }, 'Le prix est obligatoire'],
    [
      { prix: '-5' },
      'Le prix doit être un montant en euros, par exemple 12,50',
    ],
    [{ categorie: '' }, 'La catégorie est obligatoire'],
    [{ ville: ' ' }, 'La ville est obligatoire'],
    [{ codePostal: '690' }, 'Le code postal doit comporter 5 chiffres'],
    [{ codePostal: '' }, 'Le code postal doit comporter 5 chiffres'],
    [{ codePostal: '96000' }, 'Le code postal ne désigne aucun département'],
    [{ codePostal: '20900' }, 'Le code postal ne désigne aucun département'],
    [{ codePostal: '97500' }, 'Le code postal ne désigne aucun département'],
    [{ pseudo: 'a' }, 'Le pseudo doit comporter de 2 à 30 caractères'],
    [{ email: 'cycliste' }, "L'email n'est pas valide"],
    [
      { telephone: '12345' },
      'Le téléphone doit comporter 10 chiffres et commencer par 0',
    ],
  ])('rejects %o', (values, message) => {
    expect(messagesFor(values)).toEqual([message])
  })

  it.each(['01000', '20000', '20600', '95000', '97400', '97600'])(
    'accepts the code postal %s',
    (codePostal) => {
      expect(messagesFor({ codePostal })).toEqual([])
    },
  )
})

describe('rechercheSearchSchema', () => {
  it('keeps a mot-cle and a page', () => {
    expect(rechercheSearchSchema.parse({ motCle: ' velo ', page: 2 })).toEqual({
      motCle: 'velo',
      page: 2,
    })
  })

  it('reads a numeric mot-cle back as text', () => {
    expect(rechercheSearchSchema.parse({ motCle: 208 })).toEqual({
      motCle: '208',
    })
  })

  it('drops a blank mot-cle and an unusable page', () => {
    // 1e20 is past what the API reads as a page number.
    for (const page of [0, -1, 1.5, 'abc', 1e20]) {
      expect(rechercheSearchSchema.parse({ motCle: '  ', page })).toEqual({
        motCle: undefined,
        page: undefined,
      })
    }
  })

  it('keeps a known categorie and drops an unknown one', () => {
    expect(rechercheSearchSchema.parse({ categorie: 'vehicules' })).toEqual({
      categorie: 'vehicules',
    })
    expect(rechercheSearchSchema.parse({ categorie: 'licornes' })).toEqual({})
  })

  it('keeps prix bounds in euros, zero included', () => {
    expect(rechercheSearchSchema.parse({ prixMin: 0, prixMax: 12.5 })).toEqual({
      prixMin: 0,
      prixMax: 12.5,
    })
    expect(rechercheSearchSchema.parse({ prixMin: '100' })).toEqual({
      prixMin: 100,
    })
  })

  it('drops an unusable prix bound', () => {
    // 1e17 euros is past the int64 centimes the API reads.
    for (const prix of [-1, 'abc', '', 1e17]) {
      expect(rechercheSearchSchema.parse({ prixMin: prix })).toEqual({})
    }
  })

  it('drops the maximum when it is below the minimum', () => {
    expect(rechercheSearchSchema.parse({ prixMin: 500, prixMax: 100 })).toEqual(
      { prixMin: 500 },
    )
  })

  it('keeps a region or a departement code, read back as text', () => {
    expect(rechercheSearchSchema.parse({ region: 84 })).toEqual({
      region: '84',
    })
    expect(rechercheSearchSchema.parse({ departement: '2A' })).toEqual({
      departement: '2A',
    })
    expect(rechercheSearchSchema.parse({ departement: '974' })).toEqual({
      departement: '974',
    })
  })

  it('drops a malformed lieu code', () => {
    for (const code of ['', 'x', '2C', '1234', 7]) {
      expect(rechercheSearchSchema.parse({ region: code })).toEqual({})
    }
  })

  it('drops the departement when a region is set too', () => {
    expect(
      rechercheSearchSchema.parse({ region: '04', departement: '974' }),
    ).toEqual({ region: '04' })
  })

  it('keeps a ville or a code postal, read back as text', () => {
    expect(
      rechercheSearchSchema.parse({ villeOuCodePostal: ' Saint-Denis ' }),
    ).toEqual({ villeOuCodePostal: 'Saint-Denis' })
    expect(rechercheSearchSchema.parse({ villeOuCodePostal: 97400 })).toEqual({
      villeOuCodePostal: '97400',
    })
    expect(rechercheSearchSchema.parse({ villeOuCodePostal: '  ' })).toEqual({})
  })
})

describe('criteresDeRecherche', () => {
  it('turns the prix bounds from euros into centimes', () => {
    expect(
      criteresDeRecherche({
        motCle: 'velo',
        categorie: 'loisirs',
        prixMin: 0,
        prixMax: 12.5,
        page: 2,
      }),
    ).toEqual({
      motCle: 'velo',
      categorie: 'loisirs',
      prixMin: 0,
      prixMax: 1250,
      page: 2,
    })
  })

  it('leaves the absent criteria out', () => {
    expect(criteresDeRecherche({})).toEqual({})
  })

  const regions = [
    {
      code: '04',
      nom: 'La Réunion',
      departements: [{ code: '974', nom: 'La Réunion' }],
    },
  ]

  it('keeps a known lieu and the ville or code postal', () => {
    expect(
      criteresDeRecherche(
        { region: '04', villeOuCodePostal: 'Saint-Denis' },
        regions,
      ),
    ).toEqual({ region: '04', villeOuCodePostal: 'Saint-Denis' })
    expect(criteresDeRecherche({ departement: '974' }, regions)).toEqual({
      departement: '974',
    })
  })

  it('drops a region or a departement the API does not know', () => {
    expect(criteresDeRecherche({ region: '99' }, regions)).toEqual({})
    expect(criteresDeRecherche({ departement: '975' }, regions)).toEqual({})
  })
})

describe('corrigerAnnonceSchema', () => {
  const correction = {
    titre: 'Vélo de course révisé',
    description: 'Cadre carbone, photos ajoutées.',
    prix: '420,50',
    categorie: 'loisirs',
    ville: 'Lyon',
    codePostal: '69003',
  }

  it('turns the fields into the centimes the gateway takes', () => {
    expect(corrigerAnnonceSchema.parse(correction)).toEqual({
      ...correction,
      prix: 42050,
    })
  })

  it('applies the rules of the dépôt to the fields it keeps', () => {
    const result = corrigerAnnonceSchema.safeParse({
      ...correction,
      titre: 'ab',
      prix: 'beaucoup',
      codePostal: '123',
    })

    expect(
      result.success ? [] : result.error.issues.map((i) => i.message),
    ).toEqual([
      'Le titre doit comporter de 3 à 100 caractères',
      'Le prix doit être un montant en euros, par exemple 12,50',
      'Le code postal doit comporter 5 chiffres',
    ])
  })
})
