# Domain glossary

Ubiquitous language of the second-hand marketplace. One unambiguous English name per concept,
spelled as in the code, with the French word of the event storming in parentheses when it differs.
Tags: _aggregate_, _value object_, _command_, _event_, _rule_, _actor_.

An entry says what a term is, in one or two sentences, 300 characters at most. A term earns an
entry when it names an aggregate, a subdomain or bounded context, or a key concept shared by
several modules or shown to users. A setting, an option, a variant of another entry or a rule
detail gets none: it belongs to its module's `AGENTS.md` or to the code. An entry stands alone: no
reference to a spec or a file, since such links go stale.

Entries are written `- **Term** (terme français) — _tag_. Definition.` and grouped under one
heading per subdomain. Add a term the moment it is settled, in the same change as the code that
uses it; never invent a name in code that is not here.

## Domain map

| Subdomain | Type | Bounded context | Aggregates |
| --------- | ---- | --------------- | ---------- |

<!-- One row per subdomain, filled after the event storming. Type: core, supporting or generic. -->
