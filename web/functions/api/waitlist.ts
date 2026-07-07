/// <reference types="@cloudflare/workers-types" />

// POST /api/waitlist  { email, signature }
// The only data this site ever stores about a person: an opt-in email plus
// the rule id of the diagnosis they saw. Never any log content.

interface Env {
  DB: D1Database;
}

const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

export const onRequestPost: PagesFunction<Env> = async ({ request, env }) => {
  let body: { email?: string; signature?: string };
  try {
    body = await request.json();
  } catch {
    return json({ error: 'invalid JSON' }, 400);
  }

  const email = (body.email ?? '').trim().toLowerCase();
  if (!emailRe.test(email) || email.length > 254) {
    return json({ error: 'invalid email' }, 400);
  }
  const signature = (body.signature ?? '').slice(0, 120);

  await env.DB.prepare(
    'INSERT INTO waitlist (email, failure_signature) VALUES (?1, ?2) ' +
      'ON CONFLICT(email) DO UPDATE SET failure_signature = ?2',
  )
    .bind(email, signature)
    .run();

  return json({ ok: true });
};
