# Security Policy

## Reporting a vulnerability

Please report security issues privately by emailing **qinchonghanzuibang@users.noreply.github.com** with the subject `Alfrenslate security report`. Include affected version, impact, reproduction steps, and any suggested mitigation. Allow a reasonable time for acknowledgement and remediation before public disclosure.

Do not post API keys, App Secrets, authorization headers, private endpoints containing credentials, or unredacted diagnostic output in a public GitHub Issue.

If a credential may have been exposed, revoke or rotate it immediately in the provider console. Removing it from a Git commit or Issue does not make the original credential safe.

## Scope

Supported releases receive security fixes. Alfrenslate has no backend service; reports should focus on the Workflow, local storage, request construction, credential handling, packaging, or release supply chain.
