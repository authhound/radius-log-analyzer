/// <reference types="@cloudflare/workers-types" />

// POST /api/event  { name, detail }
// Anonymous funnel counters: which step of search -> analyze -> diagnosis ->
// waitlist was reached. `detail` is at most a rule id or 'no_match' — log
// content is never sent here (see the track() helper in analyzer.astro).

interface Env {
  DB: D1Database;
}

const allowedEvents = new Set([
  'analyze_clicked',
  'diagnosis_shown',
  'analyze_failed',
  'waitlist_submitted',
]);

export const onRequestPost: PagesFunction<Env> = async ({ request, env }) => {
  let body: { name?: string; detail?: string };
  try {
    body = await request.json();
  } catch {
    return new Response(null, { status: 400 });
  }

  const name = body.name ?? '';
  if (!allowedEvents.has(name)) {
    return new Response(null, { status: 400 });
  }
  const detail = (body.detail ?? '').slice(0, 120);

  await env.DB.prepare('INSERT INTO events (name, detail) VALUES (?1, ?2)')
    .bind(name, detail)
    .run();

  return new Response(null, { status: 204 });
};
