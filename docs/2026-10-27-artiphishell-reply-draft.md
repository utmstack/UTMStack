# Draft reply — Artiphishell disclosure (2026-10-27)

**Status: DRAFT — for user review before sending. Not sent. Not to be committed to a public repo without review (it documents our triage of a disclosure).**

To: Jay Vadayath (Artiphishell) — address from the thread
Subject: Re: Security disclosure : Multiple security issues in UTMStack v11

---

Hi Jay,

Thanks for the detailed reports — we've completed our independent review of all ten, against our current codebase, and I want to correct a few things on our side of the record before we close this out.

**Incident-response findings (the action and action-command update/delete endpoints).** These were written against the assumption the routes were meant to be admin-only. In UTMStack the standard user role *is* the security analyst, and authoring and executing incident-response playbooks is that role's core function — so the authorization on those routes is correct as shipped. What was not intended is that the seeded system presets (the shutdown / isolate / run-command templates and their underlying shell strings) could be modified or deleted by anyone. We've shipped enforcement that makes those immutable through the API for all roles, and we've also opened incident-job dispatch to the analyst role, which was wrongly admin-locked. Both are in our v11.3.1 cycle.

**The two SQL-injection reports on the visualization filter builder** (two of the five "critical" items describe the same code): we've reviewed the endpoints involved. Analysts have full read access to all log data by design in UTMStack, so the cross-tenant impact described does not apply to our product model. We're cleaning up the query construction as good hygiene regardless.

**The web-pdf and inputs-plugin items** both require a network position that is already inside the deployment's own trust boundary; we've evaluated them and are treating them as hardening notes rather than vulnerabilities.

**The SAML metadata fetch item** is valid as defense-in-depth, and we're folding it into our own fix cycle.

On process: we handle remediation internally and don't take paid patch engagements, so we won't be moving forward commercially — but the technical content of the disclosure was useful and we appreciate the coordination and the embargo discipline. The items above ship under our normal release process, and we'll confirm once v11.3.1 is out.

Regards,
Osmany
