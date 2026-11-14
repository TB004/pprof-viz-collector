# Security Policy

## Supported Versions

Only the latest release of `pprof-viz-collector` receives security fixes. Please make sure you
can reproduce an issue on the latest release before reporting it.

## Reporting a Vulnerability

**Do not open a public issue for security vulnerabilities.**

Report vulnerabilities privately through GitHub's private vulnerability reporting:

<https://github.com/TB004/pprof-viz-collector/security/advisories/new>

Please include:

- the affected version (module version or commit),
- a description of the issue and its impact,
- steps or a minimal program to reproduce it.

## Scope

This repository covers the `github.com/TB004/pprof-viz-collector` Go module: the ConnectRPC
service, the `Bind` and `New` entry points, bearer-token authentication, CORS, TLS/HTTP3 server
setup, and the metric collection code.

Issues in the Pprof-Viz web app belong to
[`TB004/pprof-viz-web`](https://github.com/TB004/pprof-viz-web). If you are unsure which side a
problem is on, report it here and it will be forwarded.

## Response

This project is maintained by a single developer. Reports are handled on a best-effort basis:
you can expect an acknowledgement within about 7 days. Once an issue is confirmed, a fix is
released as soon as practical and a security advisory is published crediting the reporter, unless
you prefer to remain anonymous.
