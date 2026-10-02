`htmx.min.js` is the official htmx v2.0.8 browser distribution, copied from
`https://unpkg.com/htmx.org@2.0.8/dist/htmx.min.js`. SHA-256:
`22283ef68cb7545914f0a88a1bdedc7256a703d1d580c1d255217d0a50d31313`.
It is bundled into the Go executable via `embed`; no third-party host is needed
at runtime. Its Zero-Clause BSD license is retained as `htmx.LICENSE`. The
application-specific 400 response swap policy lives in `site.js`.
