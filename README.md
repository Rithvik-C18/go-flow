# Fluxion

A React workflow editor with a Go API and PostgreSQL storage.

## Deploy the free Render demo

The root `render.yaml` provisions a Docker web service and PostgreSQL database in
Singapore. In Render, create a **Blueprint**, connect this GitHub repository,
select the deployment branch, and apply it. Render generates a JWT secret and
connects the service to its database. No local `.env` files are uploaded.

The service serves the frontend and `/api` on the same HTTPS origin. `/healthz`
checks database availability. The free database expires after 30 days; export
anything you want to keep or upgrade before expiration. Free web services can
sleep when idle.

Optional integrations require `GEMINI_API_KEY` and/or `GOOGLE_CREDENTIALS`
(service-account JSON) in the Render service environment. Core authentication,
workflow editing, HTTP nodes and conditions do not require these keys. Do not
expose shared integration credentials in a public demo without controlling who
can register and use them.

## Local development

Requirements: Go 1.26.5, Node 22.12+ and PostgreSQL.

1. Copy `.env.example` to `.env` and set `DATABASE_URL` and `JWT_SECRET`.
2. Run `go run ./cmd/go-flow`.
3. Run `npm ci` then `npm run dev` in `web/`.

Vite proxies `/api` to port 8080. Production cookies are Secure and HttpOnly;
use HTTPS when testing session refresh outside localhost. Set `ALLOWED_ORIGINS`
only if you intentionally host the frontend separately.

## Build a workflow

Create a workflow, click a node in the left panel, and select it on the canvas
to open its settings. In **HTTP Request**, paste a complete URL such as
`https://jsonplaceholder.typicode.com/todos/1`, choose the method, and add
optional query parameters, headers, and a JSON body. **Test step** runs that
node and its upstream nodes, then shows the response. A successful HTTP node
exposes `statusCode` and the fields returned by the API.

Connect an HTTP node to **Condition (If)** and compare `{{ statusCode }}` with
`200`. Use `{{ params.name }}` for JSON values passed through Test input or the
workflow run dialog. In HTTP headers, query parameters, and JSON body fields,
`{{ input.field }}` refers to the previous node's output. The expression must
occupy the entire field value. Each node also has **Advanced JSON** for options
outside its form. Saving an edited node preserves its connections.

Google Docs and Sheets forms accept the document/spreadsheet ID from between
`/d/` and `/edit` in a Google URL. Their Test step needs a Google service
account configured on the server, and Gemini needs `GEMINI_API_KEY`.

## Container

```sh
docker build -t go-flow .
docker run --rm -p 8080:8080 --env-file .env go-flow
```

The database URL must be reachable from inside the container. Database tables
are migrated on startup. Workflows belong to their creator. Existing workflows
without an owner are preserved but hidden; assign their `user_id` explicitly to
the correct account before using them. The deployment uses a fresh database and
does not copy local data or secrets.

## Checks

```sh
go test ./...
go vet ./...
cd web
npm run build
npm run lint
```
